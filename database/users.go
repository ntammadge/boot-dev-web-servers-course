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
	Id          int    `json:"id"`
	Email       string `json:"email"`
	IsChirpyRed bool   `json:"is_chirpy_red" default:"false"`
}

type internalUser struct {
	User
	Password string `json:"password"`
}

func (dbStructure *DBStructure) getUserFromEmail(email string) (intUsr *internalUser, found bool) {
	intUsr, found = dbStructure.getUser(func(intUsr internalUser) bool {
		return intUsr.Email == email
	})

	return intUsr, found
}

// Creates a user with the specified email and an incremented id
func (db *DB) CreateUser(email string, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if dbStructure.Users == nil {
		dbStructure.Users = []internalUser{}
	}

	_, found := dbStructure.getUserFromEmail(email)
	if found {
		return User{}, ErrEmailInUse
	}

	userId := len(dbStructure.Users) + 1
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	intUsr := internalUser{
		User: User{
			Id:    userId,
			Email: email},
		Password: string(hashedPassword)}
	dbStructure.Users = append(dbStructure.Users, intUsr)

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return intUsr.User, nil
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

	intrnlUser, found := dbStructure.getUserFromEmail(email)
	if !found {
		return User{}, ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(intrnlUser.Password), []byte(password))
	if err != nil {
		return User{}, err
	}

	return intrnlUser.User, nil
}

// Updates a user entry in the database based on provided values for the email and password.
//
//	Empty values will not update the corresponding field in the database.
//	Update should be authorized prior to calling this method
func (db *DB) UpdateUser(id int, email string, password string) (User, error) {
	// TODO: Reconsider how updates are performed.
	// How do updates happen when more fields are present? Keep each field update separate or execute all at once?
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	intUsr, found := dbStructure.getUserFromId(id)
	if !found {
		return User{}, ErrUserNotFound
	}

	if email != "" {
		intUsr.Email = email
	}
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, err
		}
		intUsr.Password = string(hashedPassword)
	}

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}
	return intUsr.User, nil
}

// Gets a user via a supplied selector function. The selector function defines which user field to select on
func (dbStructure *DBStructure) getUser(selector func(intUsr internalUser) bool) (intUsr *internalUser, found bool) {
	if dbStructure.Users == nil || len(dbStructure.Users) == 0 {
		return &internalUser{}, false
	}

	for i := 0; i < len(dbStructure.Users); i++ {
		if selector(dbStructure.Users[i]) {
			return &dbStructure.Users[i], true
		}
	}
	return &internalUser{}, false
}

func (dbStructure *DBStructure) getUserFromId(id int) (intUsr *internalUser, found bool) {
	intUsr, found = dbStructure.getUser(func(intUsr internalUser) bool {
		return intUsr.Id == id
	})

	return intUsr, found
}

// Upgrades a specified user to Chirpy Red
//
//	Returns the upgraded user on success. Returns an error if the database read/writer failed
func (db *DB) UpgradeUser(id int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	intUsr, found := dbStructure.getUserFromId(id)
	if !found {
		return User{}, ErrUserNotFound
	}

	intUsr.IsChirpyRed = true

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}
	return intUsr.User, nil
}
