package database

import (
	"testing"
	"time"
)

func TestRevokeTokenSuccess(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	err = testDb.RevokeToken("testing123")
	if err != nil {
		t.Fatalf("Error revoking token: %v", err)
	}
}

func TestIsTokenRevokedFound(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	revokedToken := "testing123"

	dbStructure := DBStructure{}
	dbStructure.RevokedUserTokens = map[string]time.Time{
		revokedToken: time.Now().UTC(),
	}
	err = testDb.writeDB(dbStructure)
	if err != nil {
		t.Fatalf("Error creating test database entry")
	}

	revoked, err := testDb.IsTokenRevoked(revokedToken)
	if err != nil {
		t.Fatalf("Error checking database for revoked token: %v", err)
	}
	if !revoked {
		t.Fatal("Failed to add the revoked token to the database")
	}
}

func TestIsTokenRevokedNotFound(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	dbStructure := DBStructure{}
	dbStructure.RevokedUserTokens = map[string]time.Time{
		"testing123": time.Now().UTC(),
	}
	err = testDb.writeDB(dbStructure)
	if err != nil {
		t.Fatalf("Error creating test database entry")
	}

	revoked, err := testDb.IsTokenRevoked("shouldnotbefound")
	if err != nil {
		t.Fatalf("Error checking database for revoked token: %v", err)
	}
	if revoked {
		t.Fatal("Incorrectly identified token as revoked")
	}
}

func TestRevokeTokenUpdatesDB(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	revokedToken := "testing123"

	err = testDb.RevokeToken(revokedToken)
	if err != nil {
		t.Fatalf("Error revoking token: %v", err)
	}

	revoked, err := testDb.IsTokenRevoked(revokedToken)
	if err != nil {
		t.Fatalf("Error checking database for revoked token: %v", err)
	}
	if !revoked {
		t.Fatal("Failed to update database with revoked token")
	}
}
