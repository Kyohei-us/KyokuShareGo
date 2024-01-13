package dbServices

import (
	"testing"
)

func TestFindFirstArtistByName(t *testing.T) {
	err := ConnectDB()
	if err != nil {
		t.Errorf("DB connection error; cannnot run query")
	}
	targetString := "Jump"
	artist, artistErr := FindFirstArtistByName(targetString)
	if artistErr != nil {
		t.Errorf("Find query error got artist.Name = %s; want %s", artist.Name, targetString)
	} else if artist.Name != targetString {
		t.Errorf("Wrong Artist Name returned is %s; want %s", artist.Name, targetString)
	}
}
