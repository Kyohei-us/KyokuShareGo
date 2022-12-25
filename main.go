package main

import (
	"KyokuShareGo/models"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/contrib/sessions"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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

func CreateUser(db *gorm.DB, email string, password string) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	new_user := models.User{Email: email, HashedPassword: string(hashed)}
	result := db.Create(&new_user)
	return result.Error
}

func FindUserByEmail(db *gorm.DB, email string) (models.User, error) {
	var user models.User
	result := db.First(&user, "Email = ?", email)
	return user, result.Error
}

// AuthRequired is a simple middleware to check the session
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("gin_session_username")
	if user == nil {
		// Abort the request with the appropriate error code
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	// Continue down the chain to handler etc
	c.Next()
}

func main() {
	godotenv.Load(".env")
	db_conn_string, ok := os.LookupEnv("DATABASE_URL")
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
	db.AutoMigrate(&models.User{})

	r := gin.Default()

	r.Use(sessions.Sessions("mysession", sessions.NewCookieStore([]byte("kyokusharego_secret"))))

	// Private group, require authentication to access
	private := r.Group("/session")
	private.Use(AuthRequired)
	{
		private.GET("/me", func(c *gin.Context) {
			session := sessions.Default(c)
			user := session.Get("gin_session_username")
			c.JSON(http.StatusOK, gin.H{"user": user})
		})
	}

	// ユーザー登録
	r.POST("/signup", func(c *gin.Context) {
		// バリデーション処理
		var json models.UserAuthJSONRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 登録ユーザーが重複していた場合にはじく処理
		if err := CreateUser(db, json.Email, json.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "successful",
		})
	})

	// ログイン用のhandler
	r.POST("/login", func(c *gin.Context) {
		// バリデーション処理
		var json models.UserAuthJSONRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, userFindErr := FindUserByEmail(db, json.Email)
		if userFindErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ユーザーパスワードの比較
		if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(json.Password)); err != nil {
			fmt.Println("ログインできませんでした")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else {
			fmt.Println("ログインできました")
			session := sessions.Default(c)
			session.Set("gin_session_username", user.Email)

			// c.SetCookie("gin_cookie_username", user.Email, 3600, "/", "localhost", false, true)

			if err := session.Save(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "successful",
			})
			return
		}
	})

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
