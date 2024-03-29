package apiConfig

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/trolfu/boot-dev-web-servers-course/database"
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
//
//	If an `author_id` is provided as a query parameter, the chirps are filtered to the provided author id
//	If a `sort` parameter is provided, the returned chirps will be sorted accordingly. `asc` for ascending order and `desc` for decending order. Default sort method is `asc`
func (config *apiConfig) GetChirps(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	authorIdParam := request.URL.Query().Get("author_id")
	chirps, err := config.getChirps(authorIdParam)

	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Error retrieving chirps: %v", err))
		return
	}

	sortParam := request.URL.Query().Get("sort")
	chirps = sortChirpsById(chirps, sortParam)

	respondWithSuccess(writer, http.StatusOK, chirps)
}

// Gets chirps from the database. If the author id is provided, gets chirps from that specific user
func (config *apiConfig) getChirps(authorId string) ([]database.Chirp, error) {
	id, err := strconv.Atoi(authorId)
	var chirps []database.Chirp

	if authorId == "" || err != nil {
		chirps, err = config.db.GetChirps()
	} else {
		chirps, err = config.db.GetUserChirps(id)
	}

	if err != nil {
		return nil, err
	}

	return chirps, nil
}

// Sorts chirps by their id and the provided sort order. Sort order
func sortChirpsById(chirps []database.Chirp, sortOrder string) []database.Chirp {
	ascOrder := "asc"
	descOrder := "desc"
	if sortOrder != ascOrder && sortOrder != descOrder {
		sortOrder = ascOrder
	}

	if sortOrder == ascOrder {
		slices.SortFunc(chirps, func(a, b database.Chirp) int {
			return cmp.Compare(a.Id, b.Id)
		})
	} else {
		slices.SortFunc(chirps, func(a, b database.Chirp) int {
			return cmp.Compare(b.Id, a.Id)
		})
	}
	return chirps
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
