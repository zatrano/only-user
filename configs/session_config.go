package configs

import (
	"encoding/gob"
	"time"

	"zatrano/models"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2/middleware/session"
)

var Session *session.Store

func InitSession() {
	Session = createSessionStore()
	utils.InitializeSessionStore(Session)
	utils.SLog.Info("Oturum (session) sistemi başlatıldı ve utils içinde kayıt edildi.")
}

func SetupSession() *session.Store {
	if Session == nil {
		utils.SLog.Warn("Session store isteniyor ancak henüz başlatılmamış, şimdi başlatılıyor.")
		InitSession()
	}
	return Session
}

func createSessionStore() *session.Store {
	sessionExpirationHours := utils.GetEnvAsInt("SESSION_EXPIRATION_HOURS", 24)

	cookieSecure := utils.IsProduction()

	store := session.New(session.Config{
		CookieHTTPOnly: false,
		CookieSecure:   cookieSecure,
		Expiration:     time.Duration(sessionExpirationHours) * time.Hour,
		KeyLookup:      "cookie:session_id",
		CookieSameSite: "Lax",
	})

	utils.SLog.Infof("Cookie tabanlı session sistemi %d saatlik süreyle yapılandırıldı.", sessionExpirationHours)

	registerGobTypes()

	return store
}

func registerGobTypes() {
	gob.Register(models.UserType(""))
	gob.Register(&models.User{})
	utils.SLog.Debug("Session için gob türleri kaydedildi: models.UserType, *models.User")
}
