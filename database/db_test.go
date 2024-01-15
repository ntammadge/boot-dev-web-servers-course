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
		1: {Id: 1, Body: "A chirp"},
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
		1: {Id: 1, Body: "First chirp"},
		2: {Id: 2, Body: "Another one!"},
	}}

	err = testDb.writeDB(testDBStructure)
	if err != nil {
		t.Fatalf("Error writing database: %v", err)
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
