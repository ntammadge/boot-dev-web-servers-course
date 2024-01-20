package apiConfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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

func (config *apiConfig) MiddlewareIncrementMetrics(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		config.fileserverHits++
		handler.ServeHTTP(writer, request)
	})
}

func (config *apiConfig) ApiMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if request.Method == http.MethodGet {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(fmt.Sprintf("Hits: %v", config.fileserverHits)))
		return
	}
}

func (config *apiConfig) AdminApiMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(http.StatusOK)
	// There has to be a better way of doing this
	writer.Write([]byte(
		fmt.Sprintf("<html>"+
			"<body>"+
			"<h1>Welcome, Chirpy Admin</h1>"+
			"<p>Chirpy has been visited %d times!</p>"+
			"</body>"+

			"</html>",
			config.fileserverHits)))
}

func (config *apiConfig) ResetMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	config.fileserverHits = 0
}

func (config *apiConfig) CreateChirp(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(request.Body)
	incommingChirp := database.Chirp{}
	err := decoder.Decode(&incommingChirp)

	// Check failure conditions
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	if len(incommingChirp.Body) > 140 {
		respondWithError(writer, http.StatusBadRequest, "Chirp is too long")
		return
	}

	// Valid chirp
	body := cleanChirpBody(incommingChirp.Body)
	chirp, err := config.db.CreateChirp(body)

	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error creating chirp: %v", err))
	}
	respondWithSuccess(writer, http.StatusCreated, chirp)
}

// Get a single chirp by id
func (config *apiConfig) GetChirp(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	idStr := chi.URLParam(request, "chirpId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, fmt.Sprintf("Error parsing id: %v", err))
		return
	}
	chirp, found, err := config.db.GetChirp(id)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error getting Chirps: %v", err))
		return
	}
	if !found {
		respondWithError(writer, http.StatusNotFound, fmt.Sprintf("chirp with id '%v' not found", id))
		return
	}
	respondWithSuccess(writer, http.StatusOK, chirp)
}

// Gets all Chirps
func (config *apiConfig) GetChirps(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	chirps, err := config.db.GetChirps()
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error getting Chirps: %v", err))
		return
	}
	respondWithSuccess(writer, http.StatusOK, chirps)
}

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

func cleanChirpBody(original string) string {
	// Needs to be reimplemented as a trie for large word counts and where words have overlapping leading substrings
	profaneWords := map[string]struct{}{"kerfuffle": {}, "sharbert": {}, "fornax": {}}

	originalWords := strings.Split(original, " ")
	newWords := []string{}

	for _, word := range originalWords {
		if _, found := profaneWords[strings.ToLower(word)]; found {
			newWords = append(newWords, "****")
		} else {
			newWords = append(newWords, word)
		}
	}
	return strings.Join(newWords, " ")
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
