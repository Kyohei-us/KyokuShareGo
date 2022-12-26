package main

import (
	"KyokuShareGo/controllers"
	"KyokuShareGo/dbServices"
	"os"

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

		api.GET("/kyokus", controllers.GetKyokus)
		api.POST("/kyokus", controllers.PostKyokus)

		api.GET("/artists", controllers.GetArtists)
		api.POST("/artists", controllers.PostArtists)

		api.GET("/comments", controllers.GetComments)
		api.POST("/comments", controllers.PostComments)
	}

	r.GET("/", func(c *gin.Context) {
		kyokus, err := dbServices.FindAllKyokus()

		if err != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"kyokus": kyokus,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
