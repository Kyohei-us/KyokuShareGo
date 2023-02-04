package controllers

import (
	"KyokuShareGo/dbServices"
	"KyokuShareGo/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func UserSignup(c *gin.Context) {
	// バリデーション処理
	var json models.UserAuthJSONRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 登録ユーザーが重複していた場合にはじく処理
	if err := dbServices.CreateUser(json.Email, json.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "successful",
	})
}

func UserLoginForm(c *gin.Context) {
	user_email := c.DefaultPostForm("user_email", "")
	user_password := c.DefaultPostForm("user_password", "")

	user, userFindErr := dbServices.FindUserByEmail(user_email)
	if userFindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": userFindErr.Error()})
		return
	}

	// ユーザーパスワードの比較
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(user_password)); err != nil {
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
}

func UserLogin(c *gin.Context) {
	// バリデーション処理
	var json models.UserAuthJSONRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user_email := json.Email
	user_password := json.Password

	user, userFindErr := dbServices.FindUserByEmail(user_email)
	if userFindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": userFindErr.Error()})
		return
	}

	// ユーザーパスワードの比較
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(user_password)); err != nil {
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
