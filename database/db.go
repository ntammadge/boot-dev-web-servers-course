// Defines the DB type and internal functions to interface with it

package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp           `json:"chirps"`
	Users  map[string]internalUser `json:"users"`
}

func NewDB(path string) DB {
	return DB{path: path, mux: &sync.RWMutex{}}
}

// Ensures a database file exists. If one does not exist, one is created with the minimum required JSON
func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)

	if errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(db.path)
		if err != nil {
			return err
		}

		_, err = f.WriteString("{}") // Minimum JSON required to not error while parsing
		if err != nil {
			return err
		}
		return nil
	}

	return err
}

// Gets the current database data from the file on disk
func (db *DB) loadDB() (DBStructure, error) {
	err := db.ensureDB()
	if err != nil {
		return DBStructure{}, err
	}

	db.mux.RLock()
	defer db.mux.RUnlock() // Is there a way to unlock immediately after the read is complete without needing to call multiple unlocks?

	data, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	dbStructure := DBStructure{}
	err = json.Unmarshal(data, &dbStructure)
	if err != nil {
		return DBStructure{}, err
	}

	return dbStructure, nil
}

// Writes the database structure to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	err := db.ensureDB()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(dbStructure, "", "\t")
	if err != nil {
		return err
	}

	db.mux.Lock()
	defer db.mux.Unlock()
	file, err := os.OpenFile(db.path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}
