package apiConfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Creates a chirp from the request body
func (config *apiConfig) CreateChirp(writer http.ResponseWriter, request *http.Request) {
	type chirpRequest struct {
		Body string `json:"body"`
	}

	auth := request.Header.Get("Authorization")
	if auth == "" {
		respondWithError(writer, http.StatusUnauthorized, "Missing authorization")
		return
	}
	authToken := strings.TrimPrefix(auth, "Bearer ")
	parsedToken, err := config.parseJWT(authToken)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error parsing authorization token: %v", err))
		return
	}

	writer.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(request.Body)
	incommingChirp := chirpRequest{}
	err = decoder.Decode(&incommingChirp)

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
	authorIdStr, err := parsedToken.Claims.GetSubject()
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error getting user id: %v", err))
		return
	}
	authorId, err := strconv.Atoi(authorIdStr)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error parsing user id: %v", err))
		return
	}
	chirp, err := config.db.CreateChirp(body, authorId)

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

func (config *apiConfig) DeleteChirp(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	chirpIdStr := chi.URLParam(request, "chirpId")
	if chirpIdStr == "" {
		respondWithError(writer, http.StatusBadRequest, "Unable to read chirp id from request")
		return
	}
	chirpId, err := strconv.Atoi(chirpIdStr)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error parsing chirp id: %v", err))
		return
	}

	chirp, found, err := config.db.GetChirp(chirpId)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error retrieving chirp from the database: %v", err))
		return
	}
	if !found {
		respondWithError(writer, http.StatusNotFound, "Chirp not found")
		return
	}

	auth := request.Header.Get("Authorization")
	if auth == "" {
		respondWithError(writer, http.StatusUnauthorized, "Unauthorized")
		return
	}
	authToken := strings.TrimPrefix(auth, "Bearer ")
	jwt, err := config.parseJWT(authToken)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error parsing auth token: %v", err))
		return
	}
	userIdStr, err := jwt.Claims.GetSubject()
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error reading token info: %v", err))
		return
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error parsing user id from auth token: %v", err))
		return
	}

	if chirp.AuthorId != userId {
		respondWithError(writer, http.StatusForbidden, "Not allowed")
		return
	}

	deleted, err := config.db.DeleteChirp(chirp.Id)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error deleting chirp: %v", err))
		return
	}
	if !deleted {
		respondWithError(writer, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}

	respondWithSuccess(writer, http.StatusOK, "deleted")
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
