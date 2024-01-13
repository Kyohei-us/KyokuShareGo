package dbServices

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	fmt.Println("前処理実行")

	// Read env vars from .env file
	godotenv.Load("../.env")
	db_conn_string, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		fmt.Println("ERROR: DB Connection String not found")
		return
	}
	dsn := db_conn_string

	err := ConnectDB(dsn)
	if err != nil {
		fmt.Println("DB connection error; cannnot run query")
	}

	m.Run()
}

func TestFindFirstArtistByName(t *testing.T) {
	targetString := "Kaze Fujii"
	artist, artistErr := FindFirstArtistByName(targetString)
	if artistErr != nil {
		t.Errorf("Find query error got artist.Name = %s; want %s", artist.Name, targetString)
	} else if artist.Name != targetString {
		t.Errorf("Wrong Artist Name returned is %s; want %s", artist.Name, targetString)
	}
}
