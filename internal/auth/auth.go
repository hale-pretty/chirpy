package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateJWT(secret string, userID, expiresInSeconds int) (string, error) {
	userIDstr := strconv.Itoa(userID)
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Second * time.Duration(expiresInSeconds))),
		Subject:   userIDstr,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Validate JWT
func ValidateJWT(tokenString, jwtSecret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what you expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("invalid token")
	}

	// Extract claims and userID
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	userIDStr := claims.Subject
	return userIDStr, nil
}

// Get Bearer Token
func GetBearerToken(h http.Header) (string, error) {
	tokenString := h.Get("Authorization")
	if tokenString == "" {
		return tokenString, errors.New("authorization header not found")
	}

	// Strip off "Bearer " prefix
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	if tokenString == "" {
		return tokenString, errors.New("token not found")
	}
	return tokenString, nil
}

// MakeRefreshToken makes a random 256 bit token
// encoded in hex
func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

// Get API Key
func GetAPIkey(h http.Header) (string, error) {
	APIKey := h.Get("Authorization")
	if APIKey == "" {
		return APIKey, errors.New("authorization header not found")
	}
	// Strip off "Bearer " prefix
	APIKey = strings.TrimPrefix(APIKey, "ApiKey ")
	if APIKey == "" {
		return APIKey, errors.New("token not found")
	}
	return APIKey, nil
}
