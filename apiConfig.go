package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type apiConfig struct {
	fileserverHits int
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

func postChirp(writer http.ResponseWriter, request *http.Request) {
	type incommingChirp struct {
		Body string `json:"body"`
	}
	type validChirp struct {
		Cleaned_Body string `json:"cleaned_body"`
	}

	writer.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(request.Body)
	chirp := incommingChirp{}
	err := decoder.Decode(&chirp)

	// Check failure conditions
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	if len(chirp.Body) > 140 {
		respondWithError(writer, http.StatusBadRequest, "Chirp is too long")
		return
	}

	// Valid Chirp
	cleaned_body := cleanChripBody(chirp.Body)
	writer.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(validChirp{Cleaned_Body: cleaned_body})
	writer.Write(data)
}

func respondWithError(writer http.ResponseWriter, statusCode int, errorText string) {
	type chirpError struct {
		Error string `json:"error"`
	}

	errorData, _ := json.Marshal(chirpError{Error: errorText})
	writer.WriteHeader(statusCode)
	writer.Write(errorData)
}

func cleanChripBody(original string) string {
	// Needs to be reimplemented as a trie for large word counts
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
