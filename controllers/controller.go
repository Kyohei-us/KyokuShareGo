package controllers

import (
	"KyokuShareGo/dbServices"
	"KyokuShareGo/models"
	"strconv"

	"net/http"

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

	var comments []models.Comment
	var err error

	if userId != "" {
		userIdInt, err := strconv.Atoi(userId)
		if err != nil {
			comments, _ = dbServices.FindCommentsByUserId(userIdInt)
		} else {
			comments, _ = dbServices.FindAllComments()
		}
	} else {
		comments, _ = dbServices.FindAllComments()
	}

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

	err := dbServices.CreateComment(json.KyokuId, json.UserId, json.Body)

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