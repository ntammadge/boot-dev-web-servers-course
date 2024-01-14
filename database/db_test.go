package database

import (
	"errors"
	"os"
	"testing"
)

func TestLoadEmptyDB(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	dbstructure, err := testDb.loadDB()
	if err != nil {
		t.Fatalf("Error loading the database: %v", err)
	}

	if dbstructure.Chirps != nil {
		t.Fatalf("Unexpected database structure")
	}
}

func TestLoadDB(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	testDbStructure := DBStructure{Chirps: map[int]Chirp{
		1: {Id: 1, Body: "A Chirp"},
		2: {Id: 2, Body: "ANOTHER CHIRP"},
	}}
	err = testDb.writeDB(testDbStructure)
	if err != nil {
		t.Fatalf("Error writing test data to database file: %v", err)
	}

	readDbStructure, err := testDb.loadDB()
	if err != nil {
		t.Fatalf("Error loading database: %v", err)
	}

	if len(testDbStructure.Chirps) != len(readDbStructure.Chirps) {
		t.Fatal("Mismatch in database length")
	}
	for i := 1; i <= len(testDbStructure.Chirps); i++ {
		if testDbStructure.Chirps[i].Id != readDbStructure.Chirps[i].Id ||
			testDbStructure.Chirps[i].Body != readDbStructure.Chirps[i].Body {
			t.Fatalf("Mismatch in database data for id %v", i)
		}
	}
}

func TestWriteDB(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	testDBStructure := DBStructure{Chirps: map[int]Chirp{
		1: {Id: 1, Body: "First Chirp"},
		2: {Id: 2, Body: "Another one!"},
	}}

	err = testDb.writeDB(testDBStructure)
	if err != nil {
		t.Fatalf("Error writing database: %v", err)
	}
}

func TestCreateChirp(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	chirpBody := "Something really interesting"
	chirp, err := testDb.CreateChirp(chirpBody)
	if err != nil {
		t.Fatalf("Error creating Chirp: %v", err)
	}
	if chirp.Body != chirpBody || chirp.Id != 1 {
		t.Fatal("Chirp created with incorrect data")
	}
}

func TestCreateChirpIncrementsChirpId(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	_, err = testDb.CreateChirp("First Chirp")
	if err != nil {
		t.Fatalf("Error creating first chirp: %v", err)
	}
	testChirp, err := testDb.CreateChirp("Second Chirp")
	if err != nil {
		t.Fatalf("Error creating second Chirp: %v", err)
	}

	if testChirp.Id != 2 {
		t.Fatalf("Unexpected Chirp id: %v", testChirp.Id)
	}
}

func TestCreateChirpUpdatesDatabase(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	err = testDb.writeDB(DBStructure{Chirps: map[int]Chirp{
		1: {Id: 1, Body: "First Chirp"},
		2: {Id: 2, Body: "Second Chirp"},
		3: {Id: 3, Body: "ANOTHER CHIRP"},
	}})
	if err != nil {
		t.Fatalf("Error writing initial data to database: %v", err)
	}

	testChirp, err := testDb.CreateChirp("another ANOTHER Chirp")
	if err != nil {
		t.Fatalf("Error creating test Chirp: %v", err)
	}
	if testChirp.Id != 4 {
		t.Fatal("Unexpected new Chirp id")
	}

	dbStructure, err := testDb.loadDB()
	if err != nil {
		t.Fatalf("Error loading database: %v", err)
	}
	_, found1 := dbStructure.Chirps[1]
	_, found2 := dbStructure.Chirps[2]
	_, found3 := dbStructure.Chirps[3]
	_, found4 := dbStructure.Chirps[4]

	if !found1 || !found2 || !found3 || !found4 {
		t.Fatal("Database did not contain all expected Chirps")
	}
}

func TestGetChirps(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	testDbStructure := DBStructure{Chirps: map[int]Chirp{
		1: {Id: 1, Body: "First Chirp"},
		2: {Id: 2, Body: "Second Chirp"},
		3: {Id: 3, Body: "ANOTHER CHIRP"},
	}}

	err = testDb.writeDB(testDbStructure)
	if err != nil {
		t.Fatalf("Error creating database data: %v", err)
	}

	chirps, err := testDb.GetChirps()
	if err != nil {
		t.Fatalf("Error getting Chirps: %v", err)
	}
	if len(chirps) != len(testDbStructure.Chirps) {
		t.Fatal("Unexpected number of Chirps")
	}
}

// Cleans up an existing test database file if it exists
func cleanupDbFile(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		err = os.Remove(path)
		if err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
