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
	utils.SLog.Info("Session store initialized and registered in utils")
}

func SetupSession() *session.Store {
	if Session == nil {
		utils.SLog.Warn("Session store requested but not initialized, initializing now.")
		InitSession()
	}
	return Session
}

func createSessionStore() *session.Store {
	sessionExpirationHours := utils.GetEnvAsInt("SESSION_EXPIRATION_HOURS", 24)

	store := session.New(session.Config{
		CookieHTTPOnly: false,
		CookieSecure:   utils.GetEnvWithDefault("APP_ENV", "production"),
		Expiration:     time.Duration(sessionExpirationHours) * time.Hour,
		KeyLookup:      "cookie:session_id",
		CookieSameSite: "Lax",
	})

	utils.SLog.Infof("Cookie-based session store configured with %d hour expiration", sessionExpirationHours)

	registerGobTypes()

	return store
}

func registerGobTypes() {
	gob.Register(models.UserType(""))
	gob.Register(&models.User{})
	utils.SLog.Debug("Registered gob types for session: models.UserType, *models.User")
}
