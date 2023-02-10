package controllers

import (
	"KyokuShareGo/dbServices"
	"KyokuShareGo/models"
	"fmt"
	"log"
	"strconv"

	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetKyokus(c *gin.Context) {
	artistName := c.DefaultQuery("artist_name", "")

	var kyokus []models.Kyoku
	var err error
	if artistName != "" {
		kyokus, err = dbServices.FindKyokusByArtist(artistName)
	} else {
		kyokus, err = dbServices.FindAllKyokus()
	}

	if err == nil {
		c.JSON(http.StatusOK, kyokus)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	}
}

func PostKyokus(c *gin.Context) {
	var json models.KyokuCreateJSONRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := dbServices.CreateKyokuForArtist(json.ArtistName, json.KyokuTitle)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Successful",
		})
	}

}

func GetArtists(c *gin.Context) {
	artistName := c.DefaultQuery("artist_name", "")

	var artists []models.Artist
	var err error

	if artistName != "" {
		artists, err = dbServices.FindArtistsByName(artistName)
	} else {
		artists, err = dbServices.FindAllArtists()
	}

	if err == nil {
		c.JSON(http.StatusOK, artists)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	}
}

func PostArtists(c *gin.Context) {
	var json models.ArtistCreateJSONRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := dbServices.CreateArtist(json.ArtistName)

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Successful",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	}
}

func GetComments(c *gin.Context) {
	userId := c.DefaultQuery("user_id", "")
	kyokuId := c.DefaultQuery("kyoku_id", "")

	var commentQueryString models.CommentQueryString

	if kyokuId != "" {
		kyokuIdInt, err := strconv.Atoi(kyokuId)
		if err == nil {
			commentQueryString.KyokuId = &kyokuIdInt
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "ERROR",
			})
			return
		}
	}
	if userId != "" {
		userIdInt, err := strconv.Atoi(userId)
		if err == nil {
			commentQueryString.UserId = &userIdInt
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "ERROR",
			})
			return
		}
	}

	comments, err := dbServices.FindComments(&commentQueryString)

	if err == nil {
		c.JSON(http.StatusOK, comments)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	}
}

func PostComments(c *gin.Context) {
	var json models.CommentCreateJSONRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := dbServices.CreateComment(json.KyokuID, json.UserID, json.Body)

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Successful",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	}
}

func PostCommentsLoggedIn(c *gin.Context) {
	var json models.CommentCreateLoggedInJSONRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)
	userEmail := session.Get("gin_session_username")
	user, userFindErr := dbServices.FindUserByEmail(userEmail.(string))
	if userFindErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	}

	fmt.Println("user:", user)

	err := dbServices.CreateComment(json.KyokuID, int(user.ID), json.Body)

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Successful",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	}
}

func DeleteComments(c *gin.Context) {
	commentId := c.DefaultQuery("comment_id", "")

	if commentId == "" {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
		return
	}

	commentIdInt, err := strconv.Atoi(commentId)
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
		return
	}

	var findCommentErr error
	comment, findCommentErr := dbServices.DeleteCommentById(commentIdInt)

	if findCommentErr == nil {
		c.JSON(http.StatusOK, comment)
	} else {
		log.Fatal(findCommentErr)
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	}
}

func UpdateUserDisplayName(c *gin.Context) {
	var json models.UserProfileCreateJSONRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session := sessions.Default(c)
	userEmail := session.Get("gin_session_username")
	user, userFindErr := dbServices.FindUserByEmail(userEmail.(string))
	if userFindErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	}

	_, err := dbServices.UpdateUserDisplayName(int(user.ID), json.DisplayName)

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Successful",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ERROR",
		})
	}
}
