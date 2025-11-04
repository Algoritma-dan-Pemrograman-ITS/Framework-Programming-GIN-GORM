package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hardytee1/pbkk/controllers"
	"github.com/hardytee1/pbkk/initializers"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
}

func main() {
	router := gin.Default()

	router.POST("/blog", controllers.BlogCreate)
	router.GET("/blogs", controllers.BlogGetAll)
	router.GET("/blogs/:id", controllers.BlogGetByID)
	router.PUT("/blogs/:id", controllers.BlogUpdate)
	router.DELETE("/blogs/:id", controllers.BlogSoftDelete)
	router.DELETE("/blogs/:id/hard", controllers.BlogHardDelete)

	router.Run()
}
