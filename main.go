package main

import (
	"KyokuShareGo/controllers"
	"KyokuShareGo/dbServices"
	"KyokuShareGo/models"
	"os"
	"strconv"

	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Connect to DB
	if err := dbServices.ConnectDB(); err != nil {
		return
	}

	// Set up gin router
	r := gin.Default()

	// Read env vars from .env file
	godotenv.Load(".env")
	sesh_key, KeyOk := os.LookupEnv("SESSION_KEY")
	sesh_secret, SecretOk := os.LookupEnv("SESSION_SECRET")
	if !KeyOk || !SecretOk {
		return
	}
	// Sessions を使用する宣言
	r.Use(sessions.Sessions(sesh_key, sessions.NewCookieStore([]byte(sesh_secret))))

	// CSS などの static files
	r.Static("/static", "./views/static")
	// Load HTML files in views
	r.LoadHTMLGlob("views/*.html")

	// Private group, require authentication to access
	private := r.Group("/session")
	private.Use(controllers.AuthRequired)
	{
		private.GET("/me", func(c *gin.Context) {
			session := sessions.Default(c)
			user := session.Get("gin_session_username")
			c.JSON(http.StatusOK, gin.H{"user": user})
		})
		private.PATCH("/updateDisplayName", controllers.UpdateUserDisplayName)
	}

	api := r.Group("/api")
	api.Use()
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

		// ユーザー登録
		api.POST("/signup", controllers.UserSignup)

		// ユーザーログイン
		api.POST("/login", controllers.UserLogin)
		api.POST("/login_form", controllers.UserLoginForm)

		api.GET("/kyokus", controllers.GetKyokus)
		api.GET("/kyokus/:id", controllers.GetKyokusById)
		api.POST("/kyokus", controllers.PostKyokus)

		api.GET("/artists", controllers.GetArtists)
		api.POST("/artists", controllers.PostArtists)

		api.GET("/comments", controllers.GetComments)
		api.POST("/comments", controllers.PostComments)
		api.POST("/comments_logged_in", controllers.PostCommentsLoggedIn)
		api.DELETE("/comments", controllers.DeleteComments)
	}

	r.GET("/", func(c *gin.Context) {
		kyokus, err := dbServices.FindAllKyokus()

		if err != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{})
			return
		}

		session := sessions.Default(c)
		user := session.Get("gin_session_username")
		// Not logged in
		if user == nil {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"kyokus": kyokus,
			})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"kyokus":    kyokus,
			"logged_in": true,
		})
	})

	r.GET("/kyoku/:id", func(c *gin.Context) {
		kyoku_id := c.Param("id")
		kyoku_id_int, ParseIntErr := strconv.Atoi(kyoku_id)
		if ParseIntErr != nil {
			c.HTML(http.StatusBadRequest, "kyoku_comments.html", gin.H{})
			return
		}
		kyoku, err := dbServices.FindKyokuById(kyoku_id_int)
		if err != nil {
			c.HTML(http.StatusBadRequest, "kyoku_comments.html", gin.H{})
			return
		}

		comments, err := dbServices.FindComments(&models.CommentQueryString{KyokuId: &kyoku_id_int})
		if err != nil {
			c.HTML(http.StatusBadRequest, "kyoku_comments.html", gin.H{})
			return
		}

		session := sessions.Default(c)
		user := session.Get("gin_session_username")
		if user == nil {
			c.HTML(http.StatusOK, "kyoku_comments.html", gin.H{
				"kyoku":    kyoku,
				"comments": comments,
			})
			return
		}

		c.HTML(http.StatusOK, "kyoku_comments.html", gin.H{
			"kyoku":     kyoku,
			"comments":  comments,
			"logged_in": true,
		})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})

	r.GET("/logout", controllers.UserLogout)

	r.GET("/new_comment", controllers.LoginRequired, func(c *gin.Context) {
		kyokuId := c.DefaultQuery("kyoku_id", "")
		kyokuIdInt, err := strconv.Atoi(kyokuId)
		if kyokuId != "" && err == nil {
			c.HTML(http.StatusOK, "new_comment.html", gin.H{
				"kyokuId": kyokuIdInt,
			})
			return
		}

		c.HTML(http.StatusOK, "new_comment.html", gin.H{})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
