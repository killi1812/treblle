package model

import "github.com/google/uuid"

type Session struct {
	UserId       uint      `gorm:"type:uint;unique;not null"`
	UserUuid     uuid.UUID `gorm:"type:uuid;unique;not null"`
	RefreshToken string    `gorm:"type:varchar(350);not null"`
}
