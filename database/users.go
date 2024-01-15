// Defines the User type and database functions for interacting with Users

package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailInUse   = errors.New("that email is already in use")
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

type internalUser struct {
	User
	Password string `json:"password"`
}

// Creates a user with the specified email and an incremented id
func (db *DB) CreateUser(email string, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if dbStructure.Users == nil {
		dbStructure.Users = map[string]internalUser{}
	}

	if _, found := dbStructure.Users[email]; found {
		return User{}, ErrEmailInUse
	}

	userId := len(dbStructure.Users) + 1
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	intrnlUser := internalUser{
		User: User{
			Id:    userId,
			Email: email},
		Password: string(hashedPassword)}
	dbStructure.Users[intrnlUser.Email] = intrnlUser

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return intrnlUser.User, nil
}

// Validates login credentials against a user's stored credentials in the database.
//
//	Returns the user if found and valid credentials provided.
//	Err is nil on successful validation or an error on failure
//	TODO: Should probably split the user exists validation and credentials validation for more granular response codes
func (db *DB) ValidateCredentials(email string, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if dbStructure.Users == nil {
		return User{}, ErrUserNotFound
	}

	intrnlUser, found := dbStructure.Users[email]
	if !found {
		return User{}, ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(intrnlUser.Password), []byte(password))
	if err != nil {
		return User{}, err
	}

	return intrnlUser.User, nil
}
