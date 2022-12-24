package main

import (
	"KyokuShareGo/models"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func CreateArtist(db *gorm.DB, artistName string) error {
	result := db.Create(&models.Artist{Name: artistName})
	return result.Error
}

func FindFirstArtistByName(db *gorm.DB, artistName string) (models.Artist, error) {
	var a models.Artist
	result := db.First(&a, "Name = ?", artistName)
	return a, result.Error
}

// Find any artists which name starts with artistName
func FindArtistsByName(db *gorm.DB, artistName string) ([]models.Artist, error) {
	var artists []models.Artist
	result := db.Find(&artists, "Name LIKE ?", artistName+"%")
	return artists, result.Error
}

func FindAllArtists(db *gorm.DB) ([]models.Artist, error) {
	var artists []models.Artist
	result := db.Find(&artists)
	return artists, result.Error
}

func FindKyokusByArtist(db *gorm.DB, artistName string) ([]models.Kyoku, error) {
	var a models.Artist
	tx := db.Preload("Kyokus").Begin()
	err := tx.Find(&a, "Name = ?", artistName).Commit().Error
	return a.Kyokus, err
}

func FindAllKyokus(db *gorm.DB) ([]models.Kyoku, error) {
	var kyokus []models.Kyoku
	result := db.Find(&kyokus)
	return kyokus, result.Error
}

func CreateKyokuWithoutArtist(db *gorm.DB, kyokuTitle string) (models.Kyoku, error) {
	kyoku := models.Kyoku{Title: kyokuTitle}
	result := db.Create(&kyoku)
	return kyoku, result.Error
}

func CreateKyokuForArtist(db *gorm.DB, artistName string, kyokuTitle string) error {
	artist, err := FindFirstArtistByName(db, artistName)
	if err != nil {
		return err
	}

	createError := db.Transaction(func(tx *gorm.DB) error {
		// do some database operations in the transaction (use 'tx' from this point, not 'db')
		kyoku, err := CreateKyokuWithoutArtist(tx, kyokuTitle)
		if err != nil {
			// return any error will rollback
			return err
		}

		var kk models.Kyoku
		tx.First(&kk, "Title = ?", kyoku.Title)
		artist.Kyokus = append(artist.Kyokus, kk)
		result := tx.Save(&artist)
		if result.Error != nil {
			return result.Error
		}

		// return nil will commit the whole transaction
		return nil
	})

	return createError
}

func main() {
	godotenv.Load(".env")
	db_conn_string, ok := os.LookupEnv("DATABASE_CONNECTION_STRING")
	if !ok {
		fmt.Println("ERROR: DB Connection String not found")
		return
	}
	dsn := db_conn_string
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err == nil {
		fmt.Println("DB connected!")
	} else {
		fmt.Println("ERROR: DB Connection failed")
		return
	}

	// Migrate the schema
	db.AutoMigrate(&models.Kyoku{})
	db.AutoMigrate(&models.Artist{})

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.GET("/kyokus", func(c *gin.Context) {
		artistName := c.DefaultQuery("artist_name", "")

		var kyokus []models.Kyoku
		var err error
		if artistName != "" {
			kyokus, err = FindKyokusByArtist(db, artistName)
		} else {
			kyokus, err = FindAllKyokus(db)
		}

		if err == nil {
			c.JSON(http.StatusOK, kyokus)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "ERROR",
			})
		}

	})
	// CREATE Kyoku
	r.POST("/kyokus", func(c *gin.Context) {
		var json models.KyokuCreateJSONRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := CreateKyokuForArtist(db, json.ArtistName, json.KyokuTitle)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "ERROR",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Successful",
			})
		}

	})
	r.GET("/artists", func(c *gin.Context) {
		artistName := c.DefaultQuery("artist_name", "")

		var artists []models.Artist
		var err error

		if artistName != "" {
			artists, err = FindArtistsByName(db, artistName)
		} else {
			artists, err = FindAllArtists(db)
		}

		if err == nil {
			c.JSON(http.StatusOK, artists)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "ERROR",
			})
		}
	})
	r.POST("/artists", func(c *gin.Context) {
		var json models.ArtistCreateJSONRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := CreateArtist(db, json.ArtistName)

		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "Successful",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "ERROR",
			})
		}
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
