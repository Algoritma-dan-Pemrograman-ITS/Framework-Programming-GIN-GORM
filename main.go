package main

import (
	"github.com/Algoritma-dan-Pemrograman-ITS/Framework-Programming-GIN-GORM/controllers"
	"github.com/Algoritma-dan-Pemrograman-ITS/Framework-Programming-GIN-GORM/initializers"
	"github.com/Algoritma-dan-Pemrograman-ITS/Framework-Programming-GIN-GORM/middleware"
	"github.com/Algoritma-dan-Pemrograman-ITS/Framework-Programming-GIN-GORM/models"
	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	initializers.DB.AutoMigrate(&models.User{}, &models.Blog{})
}

func main() {
	router := gin.Default()

	//apply cors
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5500")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// blog routes
	router.POST("/blog", controllers.CreateBlog)
	router.GET("/blogs", controllers.BlogGetAll)
	router.GET("/blogs/:id", controllers.BlogGetByID)
	router.PUT("/blogs/:id", controllers.BlogUpdate)
	router.DELETE("/blogs/:id", controllers.BlogSoftDelete)
	router.DELETE("/blogs/:id/hard", controllers.BlogHardDelete)

	// auth routes
	router.POST("/register", controllers.Register)
	router.POST("/login", controllers.Login)

	// protected example route
	protected := router.Group("/protected")
	protected.Use(middleware.AuthMiddleware)
	{
		protected.GET("/profile", controllers.Profile)
	}

	router.GET("/profile2", middleware.AuthMiddleware2, controllers.Profile)

	router.POST("/create-class", controllers.CreateClass)
	router.GET("/class/:id/participants", controllers.GetClassParticipants)
	router.GET("/user/:id/classes", controllers.GetClassesUserEnrolled)

	router.PUT("/class/:id/participants", controllers.UpdateClassParticipants)
	router.DELETE("/class/:id/participants/:user_id", controllers.DeleteParticipantFromClass)
	router.DELETE("/user/:id/classes", controllers.DeleteUserFromAllClasses)

	router.Run()
}
