// Defines the Chirp type and database functions for interacting with Chrips

package database

type Chirp struct {
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}

// Creates a new chirp and saves it to the database
func (db *DB) CreateChirp(body string, authorId int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	if dbStructure.Chirps == nil {
		dbStructure.Chirps = map[int]Chirp{}
	}

	id := len(dbStructure.Chirps) + 1 // TODO: Delete operation breaks this method of id determination. Need to update
	chirp := Chirp{Id: id, Body: body, AuthorId: authorId}

	dbStructure.Chirps[chirp.Id] = chirp
	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

// Gets a chirp by its id, if it exists
func (db *DB) GetChirp(id int) (chirp Chirp, found bool, err error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, false, err
	}

	if dbStructure.Chirps == nil || len(dbStructure.Chirps) == 0 {
		return Chirp{}, false, nil
	}
	chirp, found = dbStructure.Chirps[id]
	if !found {
		return Chirp{}, false, nil
	}
	return chirp, true, nil
}

// Gets all of the existing Chirps
func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	if dbStructure.Chirps == nil || len(dbStructure.Chirps) == 0 {
		return []Chirp{}, nil
	}

	chirps := []Chirp{}
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

// Gets chirps with a specific user/author id
func (db *DB) GetUserChirps(authorId int) ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	if dbStructure.Chirps == nil || len(dbStructure.Chirps) == 0 {
		return []Chirp{}, nil
	}

	chirps := []Chirp{}
	for _, chirp := range dbStructure.Chirps {
		if chirp.AuthorId == authorId {
			chirps = append(chirps, chirp)
		}
	}

	return chirps, nil
}

// Deletes a chirp from the database.
//
//	`success` is true if the chirp was removed from the database, and false if the chirp was not found, the delete failed, or an error occurred.
//	`err` is nil if the database was loaded and updated successfully, or has error information if those operations errored
func (db *DB) DeleteChirp(chirpId int) (success bool, err error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, err
	}

	if dbStructure.Chirps == nil || len(dbStructure.Chirps) == 0 {
		return false, nil
	}

	if _, found := dbStructure.Chirps[chirpId]; !found {
		return false, nil
	}

	delete(dbStructure.Chirps, chirpId)

	err = db.writeDB(dbStructure)
	if err != nil {
		return false, err
	}
	return true, nil
}
