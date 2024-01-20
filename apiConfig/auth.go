package apiConfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/trolfu/boot-dev-web-servers-course/database"
)

var (
	accessTokenIssuer          = "chirpy-access"
	accessTokenTimeoutSeconds  = 60 * 60 // 1hr
	refreshTokenIssuer         = "chirpy-refresh"
	refreshTokenTimeoutSeconds = 60 * 60 * 24 * 60 // 60 days
)

// Login a user via the request body
func (config *apiConfig) Login(writer http.ResponseWriter, request *http.Request) {
	type loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type loginResponse struct {
		database.User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	writer.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(request.Body)
	login := loginRequest{}
	err := decoder.Decode(&login)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := config.db.ValidateCredentials(login.Email, login.Password)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, err.Error())
		return
	}

	accessToken, err := config.createSignedJWT(accessTokenIssuer, accessTokenTimeoutSeconds, user.Id)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error creating access token: %v", err))
		return
	}
	refreshToken, err := config.createSignedJWT(refreshTokenIssuer, refreshTokenTimeoutSeconds, user.Id)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error creating refres token: %v", err))
		return
	}

	respondWithSuccess(writer, http.StatusOK, loginResponse{User: user, Token: accessToken, RefreshToken: refreshToken})
}

func (config *apiConfig) RefreshAuth(writer http.ResponseWriter, request *http.Request) {
	type claims struct {
		jwt.RegisteredClaims
	}
	type refreshResponse struct {
		Token string `json:"token"`
	}

	auth := request.Header.Get("Authorization")
	if auth == "" {
		respondWithError(writer, http.StatusUnauthorized, "Missing authorization")
		return
	}

	authToken := strings.TrimPrefix(auth, "Bearer ")
	jwtToken, err := jwt.ParseWithClaims(authToken, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.jwtSecret), nil
	})
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	if !jwtToken.Valid {
		respondWithError(writer, http.StatusUnauthorized, "Invalid token signature")
		return
	}
	issuer, err := jwtToken.Claims.GetIssuer()
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, err.Error())
		return
	}
	if issuer != refreshTokenIssuer {
		respondWithError(writer, http.StatusUnauthorized, "Invalid token issuer")
		return
	}
	alreadyRevoked, err := config.db.IsTokenRevoked(authToken)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error checking if the token was already revoked %v", err))
		return
	}
	if alreadyRevoked {
		respondWithError(writer, http.StatusUnauthorized, "Token previously revoked")
		return
	}

	strId, err := jwtToken.Claims.GetSubject()
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	userId, err := strconv.Atoi(strId)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	newAccessToken, err := config.createSignedJWT(accessTokenIssuer, accessTokenTimeoutSeconds, userId)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error creating new access token: %v", err))
		return
	}
	respondWithSuccess(writer, http.StatusOK, refreshResponse{newAccessToken})
}

func (config *apiConfig) RevokeAuth(writer http.ResponseWriter, request *http.Request) {
	type claims struct {
		jwt.RegisteredClaims
	}

	auth := request.Header.Get("Authorization")
	if auth == "" {
		respondWithError(writer, http.StatusUnauthorized, "Missing authorization")
		return
	}

	authToken := strings.TrimPrefix(auth, "Bearer ")
	jwtToken, err := jwt.ParseWithClaims(authToken, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.jwtSecret), nil
	})
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	issuer, err := jwtToken.Claims.GetIssuer()
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
	}
	if issuer != refreshTokenIssuer {
		respondWithError(writer, http.StatusUnauthorized, "Invalid token type")
		return
	}

	err = config.db.RevokeToken(authToken)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithSuccess(writer, http.StatusOK, nil)
}
