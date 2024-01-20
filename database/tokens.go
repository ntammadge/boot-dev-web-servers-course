package database

import "time"

// Revokes the specified refresh token and records the revokation in the database.
//
//	`token` is the plaintext refresh token from the authorization header
//	Errors if there was an error while loading or updating the database
func (db *DB) RevokeToken(token string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	if dbStructure.RevokedUserTokens == nil {
		dbStructure.RevokedUserTokens = make(map[string]time.Time)
	}

	if _, found := dbStructure.RevokedUserTokens[token]; found {
		return nil
	}

	dbStructure.RevokedUserTokens[token] = time.Now().UTC()

	err = db.writeDB(dbStructure)
	return err
}

// Checks the database to see if the refresh token has been revoked.
//
//	`token` is the plaintext refresh token from the authorization header
//	Errors if there was an error while loading the database
func (db *DB) IsTokenRevoked(token string) (bool, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, err
	}

	if dbStructure.RevokedUserTokens == nil || len(dbStructure.RevokedUserTokens) == 0 {
		return false, nil
	}

	_, found := dbStructure.RevokedUserTokens[token]
	return found, nil
}
