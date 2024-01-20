package apiConfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Create a new user
func (config *apiConfig) CreateUser(writer http.ResponseWriter, request *http.Request) {
	type userRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	writer.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(request.Body)
	incommingUser := userRequest{}
	err := decoder.Decode(&incommingUser)

	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	// Should there be logic to change status code based on error? (ex. email already used vs JSON error)
	user, err := config.db.CreateUser(incommingUser.Email, incommingUser.Password)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error creating user: %v", err))
		return
	}
	respondWithSuccess(writer, http.StatusCreated, user)
}

// Updates the user with values specified from the request
func (config *apiConfig) UpdateUser(writer http.ResponseWriter, request *http.Request) {
	type claims struct {
		jwt.RegisteredClaims
	}
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	writer.Header().Set("Content-Type", "application/json")

	auth := request.Header.Get("Authorization")
	if auth == "" {
		respondWithError(writer, http.StatusUnauthorized, "Invalid authorization")
		return
	}
	authToken := strings.TrimPrefix(auth, "Bearer ")

	jwtToken, err := jwt.ParseWithClaims(authToken, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.jwtSecret), nil
	})
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, err.Error())
		return
	}
	issuer, err := jwtToken.Claims.GetIssuer()
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, fmt.Sprintf("Error parsing token type: %v", err))
		return
	}
	if issuer != accessTokenIssuer {
		respondWithError(writer, http.StatusUnauthorized, "Invalid access token issuer")
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

	decoder := json.NewDecoder(request.Body)
	body := requestBody{}
	err = decoder.Decode(&body)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := config.db.UpdateUser(userId, body.Email, body.Password)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithSuccess(writer, http.StatusOK, user)
}
