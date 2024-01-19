package database

import "testing"

func TestCreateUser(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	userEmail := "foobar@example.com"
	expectedId := 1

	user, err := testDb.CreateUser(userEmail, "foobar")
	if err != nil {
		t.Fatalf("Error creating user: %v", err)
	}
	if user.Id != expectedId || user.Email != userEmail {
		t.Fatal("User created with incorrect data")
	}
}

func TestCreateUserUpdatesDatabase(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	createdUser, err := testDb.CreateUser("foobar@example.com", "foobar")
	if err != nil {
		t.Fatalf("Error creating user: %v", err)
	}

	dbStructure, err := testDb.loadDB()
	if err != nil {
		t.Fatalf("Error reading database: %v", err)
	}

	dbUser, found := dbStructure.getUserFromEmail(createdUser.Email)
	if !found {
		t.Fatal("Did not find new user in database")
	}
	if dbUser.Id != createdUser.Id || dbUser.Email != createdUser.Email {
		t.Fatal("User information in database differs from information at creation")
	}
}

func TestErrorIfEmailAlreadyUsed(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	createdUser, err := testDb.CreateUser("foobar@example.com", "foobar")
	if err != nil {
		t.Fatalf("Error creating user: %v", err)
	}

	_, err = testDb.CreateUser(createdUser.Email, "foobar")
	if err != ErrEmailInUse {
		t.Fatal("User creation with in use email did not fail")
	}
}

func TestValidateCredentials(t *testing.T) {
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	userEmail, userPassword := "foobar@example.com", "foobar"
	_, err = testDb.CreateUser(userEmail, userPassword)
	if err != nil {
		t.Fatalf("Error creating user: %v", err)
	}

	_, err = testDb.ValidateCredentials(userEmail, userPassword)
	if err != nil {
		t.Fatalf("Error validating user credentials: %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	targetUserId := 5
	updatedUserEmail := "updated@example.com"
	testDb := NewDB("./testdatabase.json")

	err := cleanupDbFile(testDb.path)
	if err != nil {
		t.Fatalf("Error cleaning up database file: %v", err)
	}

	err = testDb.ensureDB()
	if err != nil {
		t.Fatalf("Error creating database file: %v", err)
	}

	testDb.writeDB(DBStructure{
		Chirps: map[int]Chirp{},
		Users: []internalUser{
			{
				User: User{
					Email: "initial@example.com",
					Id:    targetUserId,
				},
				Password: "", // Password not necessary for this test
			},
		},
	})

	updatedUser, err := testDb.UpdateUser(targetUserId, updatedUserEmail, "")
	if err != nil {
		t.Fatalf("Error updating user: %v", err)
	}

	if updatedUser.Id != targetUserId || updatedUser.Email != updatedUserEmail {
		t.Fatal("User update did not update to the correct values")
	}
}
