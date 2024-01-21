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

	id := len(dbStructure.Chirps) + 1
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
