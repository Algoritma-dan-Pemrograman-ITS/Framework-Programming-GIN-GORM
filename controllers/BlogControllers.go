package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hardytee1/pbkk/initializers"
	"github.com/hardytee1/pbkk/models"
)

// func Testing (c *gin.Context) {
// 		c.JSON(200, gin.H{
// 			"message": "png",
// 		})
// 	}


func BlogCreate(c *gin.Context) {
	//get the request body
	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	
	//create a blog
	blog := models.Blog{Title: body.Title, Content: body.Content}

	result := initializers.DB.Create(&blog)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blog": blog,
	})
}

func BlogGetAll(c *gin.Context) {
	//get all blogs
	var blogs []models.Blog
	result := initializers.DB.Find(&blogs)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	//return them
	c.JSON(http.StatusOK, gin.H{
		"blogs": blogs,
	})
}

func BlogGetByID(c *gin.Context) {
	//get id from url
	id := c.Param("id")

	//find the blog with the id
	var blog models.Blog

	//get the blog
	result := initializers.DB.First(&blog, id)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	//return it
	c.JSON(http.StatusOK, gin.H{
		"blog": blog,
	})
}

func BlogUpdate(c *gin.Context) {
	//get id from url
	id := c.Param("id")

	//get request body
	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	//find the blog we want to update
	var blog models.Blog
	result := initializers.DB.First(&blog, id)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	//update the blog
	initializers.DB.Model(&blog).Updates(models.Blog{Title: body.Title, Content: body.Content})

	//return it
	c.JSON(http.StatusOK, gin.H{
		"blog": blog,
	})
}

func BlogSoftDelete(c *gin.Context) {
	//get id from url
	id := c.Param("id")

	//check if it exists
	var blog models.Blog
	resultCheck := initializers.DB.First(&blog, id)
	if resultCheck.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": resultCheck.Error.Error(),
		})
		return
	}

	//delete by id
	result := initializers.DB.Delete(&models.Blog{}, id)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}
	

	//return response
	c.JSON(http.StatusOK, gin.H{
		"message": "Blog deleted successfully",
	})
}

func BlogHardDelete(c *gin.Context) {
	//get id from url
	id := c.Param("id")

	//check if it exists
	var blog models.Blog
	resultCheck := initializers.DB.Unscoped().First(&blog, id)
	if resultCheck.Error != nil  {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": resultCheck.Error.Error(),
		})
		return
	}

	//delete by id
	result := initializers.DB.Unscoped().Delete(&models.Blog{}, id)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}
	

	//return response
	c.JSON(http.StatusOK, gin.H{
		"message": "Blog deleted successfully",
	})
}