package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/trolfu/boot-dev-web-servers-course/database"
)

type apiConfig struct {
	fileserverHits int
	db             database.DB
}

func NewAPIConfig(dbPath string) apiConfig {
	return apiConfig{fileserverHits: 0, db: database.NewDB(dbPath)}
}

func (config *apiConfig) middlewareIncrementMetrics(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		config.fileserverHits++
		handler.ServeHTTP(writer, request)
	})
}

func (config *apiConfig) apiMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if request.Method == http.MethodGet {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(fmt.Sprintf("Hits: %v", config.fileserverHits)))
		return
	}
}

func (config *apiConfig) adminApiMetrics(writer http.ResponseWriter, request *http.Request) {
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

func (config *apiConfig) resetMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	config.fileserverHits = 0
}

func (config *apiConfig) createChirp(writer http.ResponseWriter, request *http.Request) {
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
func (config *apiConfig) getChirp(writer http.ResponseWriter, request *http.Request) {
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
func (config *apiConfig) getChirps(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	chirps, err := config.db.GetChirps()
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error getting Chirps: %v", err))
		return
	}
	respondWithSuccess(writer, http.StatusOK, chirps)
}

// Create a new user
func (config *apiConfig) createUser(writer http.ResponseWriter, request *http.Request) {
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

// Login a user via the request body
func (config *apiConfig) login(writer http.ResponseWriter, request *http.Request) {
	type loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
	respondWithSuccess(writer, http.StatusOK, user)
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
