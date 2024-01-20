package apiConfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/trolfu/boot-dev-web-servers-course/database"
)

type apiConfig struct {
	fileserverHits int
	db             database.DB
	jwtSecret      string
}

func NewAPIConfig(dbPath string, jwtSecret string) apiConfig {
	return apiConfig{fileserverHits: 0, db: database.NewDB(dbPath), jwtSecret: jwtSecret}
}

func respondWithError(writer http.ResponseWriter, statusCode int, errorText string) {
	type chirpError struct {
		Error string `json:"error"`
	}

	errorData, _ := json.Marshal(chirpError{Error: errorText})
	writer.WriteHeader(statusCode)
	writer.Write(errorData)
}

func respondWithSuccess(writer http.ResponseWriter, statusCode int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error creating payload: %v", err))
		return
	}
	writer.WriteHeader(statusCode)
	writer.Write(data)
}

// Defines the process for Chirpy JWT construction
func (config *apiConfig) createSignedJWT(issuer string, timeoutSeconds int, userId int) (string, error) {
	unsignedToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(timeoutSeconds)).UTC()),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Subject:   strconv.Itoa(userId),
		},
	)
	signedToken, err := unsignedToken.SignedString([]byte(config.jwtSecret))

	if err != nil {
		return "", err
	}
	return signedToken, nil
}
