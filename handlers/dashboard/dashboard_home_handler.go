package handlers

import (
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type HomeHandler struct {
	userService services.IUserService
}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{
		userService: services.NewUserService(),
	}
}

func (h *HomeHandler) HomePage(c *fiber.Ctx) error {
	flashData, err := utils.GetFlashMessages(c)
	if err != nil {
		utils.Log.Warn("Anasayfa: Flash mesajları alınamadı", zap.Error(err))
	}

	userCount, userErr := h.userService.GetUserCount()
	if userErr != nil {
		utils.Log.Error("Anasayfa: Kullanıcı sayısı alınamadı", zap.Error(userErr))
		userCount = 0
	}

	mapData := fiber.Map{
		"Title":     "Dashboard",
		"Success":   flashData.Success,
		"Error":     flashData.Error,
		"UserCount": userCount,
	}

	return c.Render("dashboard/home/dashboard_home", mapData, "layouts/dashboard_layout")
}
