package model

import (
	"errors"
	"treblle/util/cerror"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	ROLE_USER        UserRole = "user"
	ROLE_ADMIN       UserRole = "admin"
	ROLE_SUPER_ADMIN UserRole = "superadmin"
)

// StrToUserRole converts string to UserRole
func StrToUserRole(text string) (UserRole, error) {
	role := UserRole(text)
	switch role {
	case ROLE_ADMIN:
		return ROLE_ADMIN, nil
	case ROLE_USER:
		return ROLE_USER, nil
	case ROLE_SUPER_ADMIN:
		return ROLE_SUPER_ADMIN, nil

	default:
		return "", cerror.ErrUnknownRole
	}
}

// _VALID_USER_ROLES contains all user roles and are they valid or not
var _VALID_USER_ROLES = map[UserRole]bool{
	ROLE_USER:        true,
	ROLE_SUPER_ADMIN: true,
	ROLE_ADMIN:       true,
}

type User struct {
	gorm.Model

	Uuid         uuid.UUID `gorm:"type:uuid;unique;not null"`
	Username     string    `gorm:"type:varchar(100);not null"`
	FirstName    string    `gorm:"type:varchar(100);not null"`
	LastName     string    `gorm:"type:varchar(100);not null"`
	OIB          string    `gorm:"type:char(11);unique;not null"`
	Residence    string    `gorm:"type:varchar(255);not null"`
	BirthDate    time.Time `gorm:"type:date;not null"`
	Email        string    `gorm:"type:varchar(100);unique;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	Role         UserRole  `gorm:"type:varchar(20);not null"`
	Session      *Session  `gorm:"foreignKey:UserId;null"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if _, ok := _VALID_USER_ROLES[u.Role]; !ok {
		return errors.New("invalid user role")
	}
	return nil
}

func (u *User) Update(user *User) *User {
	u.BirthDate = user.BirthDate
	u.FirstName = user.FirstName
	u.LastName = user.LastName
	u.OIB = user.OIB
	u.Residence = user.Residence
	u.Email = user.Email
	u.Role = user.Role

	return u
}
