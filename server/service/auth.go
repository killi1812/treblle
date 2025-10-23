package service

import (
	"errors"
	"treblle/app"
	"treblle/model"
	"treblle/util/auth"
	"treblle/util/cerror"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IAuthService interface {
	Login(email, password string) (string, error)
	RefreshTokens(accessToken string) (string, error)
	Logout(userUuid string) error
}

type AuthService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewAuthService() IAuthService {
	var service IAuthService

	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = &AuthService{
			db:     db,
			logger: logger,
		}
	})

	return service
}

func (s *AuthService) Login(email, password string) (string, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debugf("User not found Email = %s", email)
			return "", cerror.ErrInvalidCredentials
		}

		s.logger.Errorf("Failed to query user, error = %+v", err)
		return "", err
	}

	if !auth.VerifyPassword(user.PasswordHash, password) {
		s.logger.Debugf("Invalid password for user Email: %s, uuid: %s", user.Email, user.Uuid)
		return "", cerror.ErrInvalidCredentials
	}

	token, refresh, err := auth.GenerateTokens(&user)
	if err != nil {
		s.logger.Errorf("Failed to generate token error = %+v", err)
		return "", err
	}

	session := model.Session{}
	rez := s.db.Where("user_uuid = ?", user.Uuid).First(&session)
	if rez.Error != nil && !errors.Is(rez.Error, gorm.ErrRecordNotFound) {
		s.logger.Errorf("Failed query session, err = %w", err)
		return "", rez.Error
	} else {
		s.logger.Infof("User with uuid = %s, is logging in again", user.Uuid.String())
		s.Logout(user.Uuid.String())
	}

	session = model.Session{
		UserId:       user.ID,
		UserUuid:     user.Uuid,
		RefreshToken: refresh,
	}
	if rez := s.db.Create(&session); rez.Error != nil {
		s.logger.Errorf("Failed to create a session, err = %w", err)
		return "", rez.Error
	}

	return token, nil
}

func (s *AuthService) RefreshTokens(accessToken string) (string, error) {
	// 1. Parsing accessToken
	token, claims, err := auth.ParseToken(accessToken)
	if err != nil {
		s.logger.Errorf("Error Parsing claims err = %+v", err)
		return "", err
	}
	if !token.Valid {
		s.logger.Errorf("Token is not valid")
		return "", cerror.ErrInvalidTokenFormat
	}

	userUuid, err := uuid.Parse(claims.ID)
	if err != nil {
		s.logger.Errorf("Error Parsing uuid err = %+v", err)
		return "", err
	}

	// 2. getting and parsing refreshToken
	session := model.Session{}
	rez := s.db.Where("user_uuid = ?", userUuid).First(&session)
	if rez.Error != nil {
		s.logger.Errorf("Failed to create a session, err = %w", rez.Error)
		return "", rez.Error
	}

	var refreshClaims auth.Claims
	_, err = jwt.ParseWithClaims(session.RefreshToken, &refreshClaims, func(token *jwt.Token) (any, error) {
		return []byte(app.RefreshKey), nil
	})
	if err != nil {
		s.logger.Errorf("Error parsing refresh token claims, err = %w", err)
		return "", err
	}

	// 3. verifying token
	if claims.TokenUuid != refreshClaims.TokenUuid {
		s.logger.Errorf("Error token uuids don't match, err = %+v", err)
		return "", err
	}

	// 4. new session
	var user model.User
	rez = s.db.Where("uuid = ?", userUuid).First(&user)
	if rez.Error != nil {
		return "", rez.Error
	}

	newAccessToken, refreshToken, err := auth.GenerateTokens(&user)
	if err != nil {
		s.logger.Errorf("Failed to generate tokens, err = %w", err)
		return "", err
	}
	session.RefreshToken = refreshToken

	if rez := s.db.Where("user_uuid = ?", userUuid).Save(session); rez.Error != nil {
		return "", rez.Error
	}

	return newAccessToken, nil
}

// Logout implements IAuthService.
func (s *AuthService) Logout(userUuid string) error {
	s.logger.Debugf("logging out user with uuid = %s", userUuid)
	if rez := s.db.Where("user_uuid = ?", userUuid).Delete(&model.Session{}); rez.Error != nil {
		s.logger.Errorf("Error session: %+v", rez)
		return rez.Error
	}

	return nil
}
