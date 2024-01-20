package apiConfig

import "testing"

func TestCleanChirpBodyUnchanged(t *testing.T) {
	testCleanBody := "This is only a test"

	cleaned_Body := cleanChirpBody(testCleanBody)
	if testCleanBody != cleaned_Body {
		t.Fatalf("Expected: %s, Actual: %s\n", testCleanBody, cleaned_Body)
	}
}

func TestCleanChirpBodyCleansProfaneWord(t *testing.T) {
	testCleanBody := "This is only a kerfuffle"

	cleaned_Body := cleanChirpBody(testCleanBody)
	if testCleanBody == cleaned_Body {
		t.Fatalf("Expected: %s, Actual: %s\n", testCleanBody, cleaned_Body)
	}
}
