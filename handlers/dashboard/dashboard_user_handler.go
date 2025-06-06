package handlers

import (
	"zatrano/models"
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService services.IUserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(),
	}
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	flashData, flashErr := utils.GetFlashMessages(c)
	if flashErr != nil {
		utils.Log.Warn("Kullanıcı listesi: Flash mesajları alınamadı", zap.Error(flashErr))
	}

	var params utils.ListParams
	// QueryParser'a pointer (&) iletilmeli
	if err := c.QueryParser(&params); err != nil {
		utils.Log.Warn("Kullanıcı listesi: Query parametreleri parse edilemedi, varsayılanlar kullanılıyor.", zap.Error(err))
		params = utils.ListParams{
			Page: utils.DefaultPage, PerPage: utils.DefaultPerPage,
			SortBy: utils.DefaultSortBy, OrderBy: utils.DefaultOrderBy,
		}
	}

	if params.Page <= 0 {
		params.Page = utils.DefaultPage
	}
	if params.PerPage <= 0 {
		params.PerPage = utils.DefaultPerPage
	} else if params.PerPage > utils.MaxPerPage {
		utils.Log.Warn("Sayfa başına istenen kayıt sayısı limiti aştı, varsayılana çekildi.",
			zap.Int("requested", params.PerPage), zap.Int("max", utils.MaxPerPage), zap.Int("default", utils.DefaultPerPage))
		params.PerPage = utils.DefaultPerPage
	}
	if params.SortBy == "" {
		params.SortBy = utils.DefaultSortBy
	}
	if params.OrderBy == "" {
		params.OrderBy = utils.DefaultOrderBy
	}

	paginatedResult, dbErr := h.userService.GetAllUsersPaginated(params)

	renderData := fiber.Map{
		"Title":     "Kullanıcılar",
		"CsrfToken": c.Locals("csrf"),
		"Result":    paginatedResult,
		"Params":    params,
		"Success":   flashData.Success,
		"Error":     flashData.Error,
	}

	if dbErr != nil {
		dbErrMsg := "Kullanıcılar getirilirken bir hata oluştu."
		utils.Log.Error("Kullanıcı listesi DB Hatası", zap.Error(dbErr))
		if existingErr, ok := renderData["Error"].(string); ok && existingErr != "" {
			renderData["Error"] = existingErr + " | " + dbErrMsg
		} else {
			renderData["Error"] = dbErrMsg
		}
		renderData["Result"] = &utils.PaginatedResult{
			Data: []models.User{},
			Meta: utils.PaginationMeta{
				CurrentPage: params.Page, PerPage: params.PerPage, TotalItems: 0, TotalPages: 0,
			},
		}
	}

	return c.Render("dashboard/users/dashboard_users_list", renderData, "layouts/dashboard_layout")
}

func (h *UserHandler) ShowCreateUser(c *fiber.Ctx) error {
	currentError := ""

	flashData, flashErr := utils.GetFlashMessages(c)
	if flashErr != nil {
		utils.Log.Warn("Kullanıcı oluşturma formu: Flash mesajları alınamadı", zap.Error(flashErr))
	}

	mapData := fiber.Map{
		"Title":     "Yeni Kullanıcı Ekle",
		"CsrfToken": c.Locals("csrf"),
		"Success":   flashData.Success,
	}

	combinedError := flashData.Error
	if currentError != "" {
		if combinedError != "" {
			combinedError += " | " + currentError
		} else {
			combinedError = currentError
		}
	}
	if combinedError != "" {
		mapData["Error"] = combinedError
	}

	return c.Render("dashboard/users/dashboard_users_create", mapData, "layouts/dashboard_layout")
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	type Request struct {
		Name     string `form:"name"`
		Account  string `form:"account"`
		Password string `form:"password"`
		Status   string `form:"status"`
		Type     string `form:"type"`
		TeamID   string `form:"team_id"`
	}
	var req Request

	renderError := func(errorMsg string, statusCode int, formData Request) error {
		mapData := fiber.Map{
			"Title":     "Yeni Kullanıcı Ekle",
			"CsrfToken": c.Locals("csrf"),
			"Error":     errorMsg,
			"FormData":  formData,
		}
		return c.Status(statusCode).Render("dashboard/users/dashboard_users_create", mapData, "layouts/dashboard_layout")
	}

	if err := c.BodyParser(&req); err != nil {
		utils.SLog.Warnf("Kullanıcı oluşturma isteği ayrıştırılamadı: %v", err)
		return renderError("Geçersiz veri formatı veya eksik alanlar.", fiber.StatusBadRequest, req)
	}

	if req.Name == "" || req.Account == "" || req.Password == "" || req.Type == "" {
		return renderError("Ad, Hesap Adı, Şifre ve Kullanıcı Tipi alanları zorunludur.", fiber.StatusBadRequest, req)
	}

	status := req.Status == "true"

	user := models.User{
		Name:     req.Name,
		Account:  req.Account,
		Password: req.Password,
		Status:   status,
		Type:     models.UserType(req.Type),
	}

	if err := h.userService.CreateUser(&user); err != nil {
		utils.Log.Error("Kullanıcı oluşturulamadı (Servis Hatası)", zap.String("account", req.Account), zap.Error(err))
		return renderError("Kullanıcı oluşturulamadı: "+err.Error(), fiber.StatusInternalServerError, req)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Kullanıcı başarıyla oluşturuldu.")
	return c.Redirect("/dashboard/users", fiber.StatusFound)
}

func (h *UserHandler) ShowUpdateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		utils.Log.Warn("Kullanıcı güncelleme formu: Geçersiz ID parametresi", zap.String("param", c.Params("id")))
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz kullanıcı ID'si.")
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}
	userID := uint(id)

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		var errMsg string
		if err == services.ErrUserServiceUserNotFound {
			utils.Log.Warn("Kullanıcı güncelleme formu: Kullanıcı bulunamadı", zap.Uint("user_id", userID))
			errMsg = "Düzenlenecek kullanıcı bulunamadı."
		} else {
			utils.Log.Error("Kullanıcı güncelleme formu: Kullanıcı alınamadı (Servis Hatası)", zap.Uint("user_id", userID), zap.Error(err))
			errMsg = "Kullanıcı bilgileri alınırken hata oluştu."
		}
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, errMsg)
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}

	currentError := ""

	flashData, flashErr := utils.GetFlashMessages(c)
	if flashErr != nil {
		utils.Log.Warn("Kullanıcı güncelleme formu: Flash mesajları alınamadı", zap.Uint("user_id", userID), zap.Error(flashErr))
	}

	mapData := fiber.Map{
		"Title":     "Kullanıcı Düzenle",
		"User":      user,
		"CsrfToken": c.Locals("csrf"),
		"Success":   flashData.Success,
	}

	combinedError := flashData.Error
	if currentError != "" {
		if combinedError != "" {
			combinedError += " | " + currentError
		} else {
			combinedError = currentError
		}
	}
	if combinedError != "" {
		mapData["Error"] = combinedError
	}

	return c.Render("dashboard/users/dashboard_users_update", mapData, "layouts/dashboard_layout")
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		utils.Log.Warn("Kullanıcı güncelleme: Geçersiz ID parametresi", zap.String("param", c.Params("id")))
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz kullanıcı ID'si.")
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}
	userID := uint(id)
	redirectPathOnSuccess := "/dashboard/users"

	type Request struct {
		Name     string `form:"name"`
		Account  string `form:"account"`
		Password string `form:"password"`
		Status   string `form:"status"`
		Type     string `form:"type"`
	}
	var req Request

	renderError := func(errorMsg string, statusCode int, formData Request) error {
		user, _ := h.userService.GetUserByID(userID)
		mapData := fiber.Map{
			"Title":     "Kullanıcı Düzenle",
			"CsrfToken": c.Locals("csrf"),
			"Error":     errorMsg,
			"User":      user,
			"FormData":  formData,
		}
		return c.Status(statusCode).Render("dashboard/users/dashboard_users_update", mapData, "layouts/dashboard_layout")
	}

	if err := c.BodyParser(&req); err != nil {
		utils.Log.Warn("Kullanıcı güncelleme: Form verileri okunamadı", zap.Uint("user_id", userID), zap.Error(err))
		return renderError("Form verileri okunamadı veya eksik.", fiber.StatusBadRequest, req)
	}

	if req.Name == "" || req.Account == "" || req.Type == "" {
		return renderError("Ad, Hesap Adı ve Kullanıcı Tipi alanları zorunludur.", fiber.StatusBadRequest, req)
	}

	status := req.Status == "true"

	userUpdateData := &models.User{
		Name:    req.Name,
		Account: req.Account,
		Status:  status,
		Type:    models.UserType(req.Type),
	}
	if req.Password != "" {
		userUpdateData.Password = req.Password
	}

	if err := h.userService.UpdateUser(userID, userUpdateData); err != nil {
		errMsg := "Kullanıcı güncellenemedi: " + err.Error()
		statusCode := fiber.StatusInternalServerError

		if err == services.ErrUserServiceUserNotFound {
			utils.Log.Warn("Kullanıcı güncelleme: Kullanıcı bulunamadı (Servis hatası)", zap.Uint("user_id", userID))
			errMsg = "Güncellenecek kullanıcı bulunamadı."
			_ = utils.SetFlashMessage(c, utils.FlashErrorKey, errMsg)
			return c.Redirect(redirectPathOnSuccess, fiber.StatusSeeOther)
		} else if _, ok := err.(models.ModelError); ok {
			statusCode = fiber.StatusBadRequest
		} else if err == services.ErrPasswordUpdateFailed || err == services.ErrPasswordHashingFailed {
			statusCode = fiber.StatusBadRequest
		}

		utils.Log.Error("Kullanıcı güncelleme: Handler'da servis hatası yakalandı", zap.Uint("user_id", userID), zap.Error(err))
		return renderError(errMsg, statusCode, req)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Kullanıcı başarıyla güncellendi.")
	return c.Redirect(redirectPathOnSuccess, fiber.StatusFound)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		utils.Log.Warn("Kullanıcı silme: Geçersiz ID parametresi", zap.String("param", c.Params("id")))
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz kullanıcı ID'si.")
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}
	userID := uint(id)

	if err := h.userService.DeleteUser(userID); err != nil {
		var errMsg string
		if err == services.ErrUserServiceUserNotFound {
			utils.Log.Warn("Kullanıcı silme: Kullanıcı bulunamadı", zap.Uint("user_id", userID))
			errMsg = "Silinecek kullanıcı bulunamadı."
		} else {
			utils.Log.Error("Kullanıcı silme: Servis hatası", zap.Uint("user_id", userID), zap.Error(err))
			errMsg = "Kullanıcı silinemedi: " + err.Error()
		}
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, errMsg)
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Kullanıcı başarıyla silindi.")
	return c.Redirect("/dashboard/users", fiber.StatusFound)
}
