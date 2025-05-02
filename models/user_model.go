package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type ModelError string

func (e ModelError) Error() string {
	return string(e)
}

const (
	ErrInvalidUserType        ModelError = "geçersiz kullanıcı tipi (UserType)"
	ErrPasswordCannotBeEmpty  ModelError = "şifre boş olamaz"
	ErrInvalidUpdateTypeField ModelError = "güncelleme verisinde geçersiz 'type' alanı tipi" // Update için
)

type UserType string

const (
	System UserType = "system"
	Panel  UserType = "panel"
)

func (UserType) GormDataType() string {
	return "user_type"
}
func (UserType) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	if db.Dialector.Name() == "postgres" {
		return "user_type"
	}
	return "varchar(10)"
}

type User struct {
	gorm.Model
	Name     string   `gorm:"size:100;not null;index"`
	Account  string   `gorm:"size:100;unique;not null"`
	Password string   `gorm:"size:255;not null"`
	Status   bool     `gorm:"default:true;index"`
	Type     UserType `gorm:"type:user_type;not null;default:'panel';index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Password == "" {
		return ErrPasswordCannotBeEmpty
	}
	hashed, bcryptErr := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if bcryptErr != nil {
		return bcryptErr
	}
	u.Password = string(hashed)

	validTypes := map[UserType]bool{System: true, Panel: true}
	if _, typeIsValid := validTypes[u.Type]; !typeIsValid {
		return ErrInvalidUserType
	}

	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	var userType UserType
	knownType := false
	currentUserType := u.Type

	if tx.Statement.Dest != nil {
		if destMap, ok := tx.Statement.Dest.(map[string]interface{}); ok {
			if t, exists := destMap["type"]; exists {
				knownType = true
				if typeStr, okStr := t.(string); okStr {
					userType = UserType(typeStr)
				} else if typeVal, okType := t.(UserType); okType {
					userType = typeVal
				} else {
					return ErrInvalidUpdateTypeField
				}
			}
		}
	}

	if !knownType {
		userType = currentUserType
	}

	validTypes := map[UserType]bool{System: true, Panel: true}
	if _, typeIsValid := validTypes[userType]; !typeIsValid {
		if string(userType) != "" {
			return ErrInvalidUserType
		}
	}

	return nil
}

func (u *User) SetPassword(password string) error {
	if password == "" {
		return ErrPasswordCannotBeEmpty
	}
	hashed, bcryptErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if bcryptErr != nil {
		return bcryptErr
	}
	u.Password = string(hashed)
	return nil
}
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
