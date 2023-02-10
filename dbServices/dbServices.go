package dbServices

import (
	"KyokuShareGo/models"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func ConnectDB() error {
	godotenv.Load(".env")
	db_conn_string, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		fmt.Println("ERROR: DB Connection String not found")
		err := fmt.Errorf("%s: %s", "ERROR", "DB Connection String not found")
		return err
	}
	dsn := db_conn_string
	db_connection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err == nil {
		fmt.Println("DB connected!")
	} else {
		fmt.Println("ERROR: DB Connection failed")
		return err
	}

	db = db_connection

	// Migrate the schema
	db.AutoMigrate(&models.Kyoku{})
	db.AutoMigrate(&models.Artist{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Comment{})

	return nil
}

func CreateArtist(artistName string) error {
	result := db.Create(&models.Artist{Name: artistName})
	return result.Error
}

func FindFirstArtistByName(artistName string) (models.Artist, error) {
	var a models.Artist
	result := db.First(&a, "Name = ?", artistName)
	return a, result.Error
}

// Find any artists which name starts with artistName
func FindArtistsByName(artistName string) ([]models.Artist, error) {
	var artists []models.Artist
	result := db.Find(&artists, "Name LIKE ?", artistName+"%")
	return artists, result.Error
}

func FindAllArtists() ([]models.Artist, error) {
	var artists []models.Artist
	result := db.Find(&artists)
	return artists, result.Error
}

func FindKyokusByArtist(artistName string) ([]models.Kyoku, error) {
	var a models.Artist
	tx := db.Preload("Kyokus").Begin()
	err := tx.Find(&a, "Name = ?", artistName).Commit().Error
	return a.Kyokus, err
}

func FindAllKyokus() ([]models.Kyoku, error) {
	kyokus := []models.Kyoku{}
	result := db.Preload("Artists").Find(&kyokus)
	return kyokus, result.Error
}

func FindKyokuById(kyoku_id int) (models.Kyoku, error) {
	var kyoku models.Kyoku
	result := db.Preload("Artists").First(&kyoku, "Id = ?", kyoku_id)
	return kyoku, result.Error
}

func createKyokuWithoutArtist(db *gorm.DB, kyokuTitle string) (models.Kyoku, error) {
	kyoku := models.Kyoku{Title: kyokuTitle}
	result := db.Create(&kyoku)
	return kyoku, result.Error
}

func CreateKyokuForArtist(artistName string, kyokuTitle string) error {
	artist, err := FindFirstArtistByName(artistName)
	if err != nil {
		return err
	}

	createError := db.Transaction(func(tx *gorm.DB) error {
		// do some database operations in the transaction (use 'tx' from this point, not 'db')
		kyoku, err := createKyokuWithoutArtist(tx, kyokuTitle)
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

func CreateUser(email string, password string) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	new_user := models.User{Email: email, HashedPassword: string(hashed)}
	result := db.Create(&new_user)
	return result.Error
}

func FindUserByEmail(email string) (models.User, error) {
	var user models.User
	result := db.First(&user, "Email = ?", email)
	return user, result.Error
}

func UpdateUserDisplayName(userId int, displayName string) (models.User, error) {
	var user models.User
	result := db.First(&user, "ID = ?", userId)
	if result.Error != nil {
		return models.User{}, result.Error
	}
	user.DisplayName = displayName
	saveResult := db.Save(&user)
	return user, saveResult.Error
}

func CreateComment(kyokuId int, userId int, body string) error {
	createError := db.Transaction(func(tx *gorm.DB) error {
		// do some database operations in the transaction (use 'tx' from this point, not 'db')
		comment := models.Comment{Body: body, KyokuID: kyokuId, UserID: userId}
		result := tx.Create(&comment)
		if result.Error != nil {
			// return any error will rollback
			return result.Error
		}

		// return nil will commit the whole transaction
		return nil
	})

	return createError
}

func FindCommentsByUserId(userId int) ([]models.Comment, error) {
	var comments []models.Comment
	result := db.Find(&comments, "user_id = ?", strconv.Itoa(userId))
	return comments, result.Error
}

// func FindCommentsByKyokuId(kyokuId int) ([]models.Comment, error) {
// 	commentList := []models.Comment{}
// 	rows, err := db.Table("comments").Select("comments.body, comments.id, comments.created_at, comments.updated_at, comments.deleted_at, comments.kyoku_id, comments.user_id").Where("comments.kyoku_id = ? AND comments.deleted_at IS ?", kyokuId, gorm.Expr("NULL")).Rows()
// 	if err == nil {
// 		defer rows.Close()
// 		if rows.Next() {
// 			db.ScanRows(rows, &commentList)
// 		}
// 	}
// 	return commentList, nil
// }

// func FindCommentById(commentId int) (models.Comment, error) {
// 	comment := models.Comment{}
// 	result := db.First(&comment, "ID = ?", strconv.Itoa(commentId))
// 	return comment, result.Error
// }

func DeleteCommentById(commentId int) (bool, error) {
	comment := models.Comment{}
	_ = db.First(&comment, "ID = ?", strconv.Itoa(commentId))
	result := db.Delete(&comment)
	if result.Error == nil {
		return true, result.Error
	} else {
		return false, result.Error
	}
}

// func FindAllComments() ([]models.Comment, error) {
// 	var comments []models.Comment
// 	result := db.Find(&comments)
// 	return comments, result.Error
// }

func FindComments(commentQueryString *models.CommentQueryString) ([]models.Comment, error) {
	whereClause := ""
	if commentQueryString.KyokuId != nil {
		whereClause = whereClause + "kyoku_id = ?"
		if commentQueryString.UserId != nil {
			whereClause = whereClause + " AND user_id = ?"
			whereClause = whereClause + " AND deleted_at IS ?"
		} else {
			whereClause = whereClause + " AND deleted_at IS ?"
		}
	} else if commentQueryString.UserId != nil {
		whereClause = whereClause + "user_id = ?"
		whereClause = whereClause + " AND deleted_at IS ?"
	} else {
		whereClause = whereClause + "deleted_at IS ?"
	}

	var comments []models.Comment
	var result *gorm.DB
	if commentQueryString.KyokuId != nil {
		if commentQueryString.UserId != nil {
			result = db.Preload("User").Preload("Kyoku").Where(whereClause, commentQueryString.KyokuId, commentQueryString.UserId, gorm.Expr("NULL")).Find(&comments)
		} else {
			result = db.Preload("User").Preload("Kyoku").Where(whereClause, commentQueryString.KyokuId, gorm.Expr("NULL")).Find(&comments)
		}
	} else if commentQueryString.UserId != nil {
		result = db.Preload("User").Preload("Kyoku").Where(whereClause, commentQueryString.UserId, gorm.Expr("NULL")).Find(&comments)
	} else {
		result = db.Preload("User").Preload("Kyoku").Where(whereClause, gorm.Expr("NULL")).Find(&comments)
	}
	if result.Error != nil {
		return []models.Comment{}, result.Error
	}
	return comments, nil
}
