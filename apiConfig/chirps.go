package apiConfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/trolfu/boot-dev-web-servers-course/database"
)

// Creates a chirp from the request body
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
