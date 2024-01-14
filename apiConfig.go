package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/trolfu/boot-dev-web-servers-course/database"
)

type apiConfig struct {
	fileserverHits int
	db             database.DB
}

func NewAPIConfig() apiConfig {
	return apiConfig{fileserverHits: 0, db: database.NewDB("./database.json")}
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

	// Valid Chirp
	body := cleanChirpBody(incommingChirp.Body)
	chirp, err := config.db.CreateChirp(body)

	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error creating Chirp: %v", err))
	}
	respondWithSuccess(writer, http.StatusCreated, chirp)
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
