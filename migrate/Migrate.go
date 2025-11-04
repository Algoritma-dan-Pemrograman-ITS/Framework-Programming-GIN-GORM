package main

import (
	"github.com/hardytee1/pbkk/initializers"
	"github.com/hardytee1/pbkk/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
}

func main() {
	initializers.DB.AutoMigrate(&models.Blog{})
}	