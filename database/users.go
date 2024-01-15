// Defines the User type and database functions for interacting with Users

package database

import "errors"

var (
	ErrEmailInUse = errors.New("that email is already in use")
)

type User struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

// Creates a user with the specified email and an incremented id
func (db *DB) CreateUser(email string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if dbStructure.Users == nil {
		dbStructure.Users = map[string]User{}
	}

	if _, found := dbStructure.Users[email]; found {
		return User{}, ErrEmailInUse
	}

	userId := len(dbStructure.Users) + 1
	user := User{Id: userId, Email: email}
	dbStructure.Users[user.Email] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
