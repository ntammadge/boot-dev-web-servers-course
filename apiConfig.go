package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	type chirpError struct {
		Error string `json:"error"`
	}
	type validChirp struct {
		Valid bool `json:"valid"`
	}

	writer.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(request.Body)
	chirp := incommingChirp{}
	err := decoder.Decode(&chirp)

	// Check failure conditions
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		data, _ := json.Marshal(chirpError{Error: err.Error()})
		writer.Write(data)
		return
	}
	if len(chirp.Body) > 140 {
		writer.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(chirpError{Error: "Chirp is too long"})
		writer.Write(data)
		return
	}

	// Valid Chirp
	writer.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(validChirp{Valid: true})
	writer.Write(data)
}
