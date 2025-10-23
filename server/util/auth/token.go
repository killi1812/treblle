package auth

import (
	"treblle/app"
	"treblle/model"
	"treblle/util/cerror"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Claims struct {
	jwt.RegisteredClaims
	Email     string         `json:"email"`
	Username  string         `json:"username"`
	Role      model.UserRole `json:"role"`
	TokenUuid uuid.UUID      `json:"uuid"`
}

const (
	_ACCESS_TOKEN_DURATION  = 5 * time.Minute
	_REFRESH_TOKEN_DURATION = 7 * 24 * time.Hour
)

func ParseToken(authHeader string) (*jwt.Token, *Claims, error) {
	// Parse token
	if len(authHeader) <= len("Bearer ") || authHeader[:len("Bearer ")] != "Bearer " {
		zap.S().Debugf("token: %s", authHeader)
		return nil, nil, cerror.ErrInvalidTokenFormat
	}
	tokenString := authHeader[len("Bearer "):]
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(app.AccessKey), nil
	})
	if err != nil {
		return nil, nil, err
	}

	return token, &claims, nil
}

// GenerateTokens return a jwt access token and refresh token or an error
func GenerateTokens(user *model.User) (string, string, error) {
	if user == nil {
		return "", "", cerror.ErrUserIsNil
	}
	uuidPair := uuid.New()
	accessTokenClaims := &Claims{
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		TokenUuid: uuidPair,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(_ACCESS_TOKEN_DURATION)),
			ID:        user.Uuid.String(),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(app.AccessKey))
	if err != nil {
		zap.S().Errorf("Failed to generate access token err = %w", err)
		return "", "", err
	}

	refreshTokenClaims := &Claims{
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		TokenUuid: uuidPair,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(_REFRESH_TOKEN_DURATION)),
			ID:        user.Uuid.String(),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(app.RefreshKey))
	if err != nil {
		zap.S().Errorf("Failed to generate refresh token err = %w", err)
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
