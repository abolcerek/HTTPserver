package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	token_claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy-access",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
		Subject: userID.String(),
	})
	token, err := token_claim.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return token, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != "chirpy-access" {
		return uuid.Nil, err
	}
	string_id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(string_id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	bearer_token := headers.Get("Authorization")
	if len(bearer_token) == 0 {
		return "", fmt.Errorf("No bearer token provided")
	}
	split_string := strings.TrimPrefix(bearer_token, "Bearer")
	token_string := strings.TrimSpace(split_string)
	if len(token_string) == 0 {
		return "", fmt.Errorf("No bearer token provided")
	}
	return token_string, nil
}

func MakeRefreshToken() string {
	data := make([]byte, 32)
	rand.Read(data)
	hex_string := hex.EncodeToString(data)
	return hex_string
}

func GetAPIKey(headers http.Header) (string, error) {
	ApiKey := headers.Get("Authorization")
	if len(ApiKey) == 0 {
		return "", fmt.Errorf("No API key provided")
	}
	split_string := strings.TrimPrefix(ApiKey, "ApiKey")
	api_string := strings.TrimSpace(split_string)
	if len(api_string) == 0 {
		return "", fmt.Errorf("No Api Key provided")
	}
	return api_string, nil
}