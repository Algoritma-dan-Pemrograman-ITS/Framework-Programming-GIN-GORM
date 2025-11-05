package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hardytee1/pbkk/controllers"
	"github.com/hardytee1/pbkk/initializers"
	"github.com/hardytee1/pbkk/middleware"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
}

func main() {
	router := gin.Default()

	// blog routes
	router.POST("/blog", controllers.BlogCreate)
	router.GET("/blogs", controllers.BlogGetAll)
	router.GET("/blogs/:id", controllers.BlogGetByID)
	router.PUT("/blogs/:id", controllers.BlogUpdate)
	router.DELETE("/blogs/:id", controllers.BlogSoftDelete)
	router.DELETE("/blogs/:id/hard", controllers.BlogHardDelete)

	// auth routes
	router.POST("/signup", controllers.Signup)
	router.POST("/signin", controllers.Signin)

	// protected example route
	protected := router.Group("/protected")
	protected.Use(middleware.AuthMiddleware)
	{
		protected.GET("/profile", controllers.Profile)
	}

	router.GET("/profile2", middleware.AuthMiddleware2, controllers.Profile)

	router.Run()
}
