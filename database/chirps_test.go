package database

import "testing"

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
	chirpAuthorId := 5
	chirp, err := testDb.CreateChirp(chirpBody, chirpAuthorId)
	if err != nil {
		t.Fatalf("Error creating chirp: %v", err)
	}
	if chirp.Body != chirpBody || chirp.Id != 1 || chirp.AuthorId != chirpAuthorId {
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

	_, err = testDb.CreateChirp("First chirp", 10)
	if err != nil {
		t.Fatalf("Error creating first chirp: %v", err)
	}
	testChirp, err := testDb.CreateChirp("Second chirp", 50)
	if err != nil {
		t.Fatalf("Error creating second chirp: %v", err)
	}

	if testChirp.Id != 2 {
		t.Fatalf("Unexpected chirp id: %v", testChirp.Id)
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
		1: {Id: 1, Body: "First chirp", AuthorId: 5},
		2: {Id: 2, Body: "Second chirp", AuthorId: 7},
		3: {Id: 3, Body: "ANOTHER CHIRP", AuthorId: 3},
	}})
	if err != nil {
		t.Fatalf("Error writing initial data to database: %v", err)
	}

	testChirp, err := testDb.CreateChirp("another ANOTHER chirp", 33)
	if err != nil {
		t.Fatalf("Error creating test chirp: %v", err)
	}
	if testChirp.Id != 4 {
		t.Fatal("Unexpected new chirp id")
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
		1: {Id: 1, Body: "First chirp"},
		2: {Id: 2, Body: "Second chirp"},
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

func TestGetChirp(t *testing.T) {
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
		1: {Id: 1, Body: "First chirp"},
		2: {Id: 2, Body: "Second chirp"},
		3: {Id: 3, Body: "ANOTHER CHIRP"},
	}}

	err = testDb.writeDB(testDbStructure)
	if err != nil {
		t.Fatalf("Error creating database data: %v", err)
	}

	targetChirpId := 3

	chirp, found, err := testDb.GetChirp(targetChirpId)
	if err != nil {
		t.Fatalf("Error getting chirp: %v", err)
	}
	if !found {
		t.Fatal("Chirp not found")
	}
	if chirp.Id != targetChirpId {
		t.Fatalf("Expected chirp '%v'. Actual chirp '%v'", targetChirpId, chirp.Id)
	}
}
