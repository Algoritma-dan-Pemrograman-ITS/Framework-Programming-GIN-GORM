package controllers

import (
	"net/http"
	"strconv"

	"github.com/Algoritma-dan-Pemrograman-ITS/Framework-Programming-GIN-GORM/initializers"
	"github.com/Algoritma-dan-Pemrograman-ITS/Framework-Programming-GIN-GORM/models"
	"github.com/gin-gonic/gin"
)

func CreateClass(c *gin.Context) {
	// parse request body
	var body struct {
		ClassName string `json:"class_name" binding:"required"`
		ClassCode string `json:"class_code" binding:"required"`
		UserIDs   []uint `json:"user_ids" binding:"required"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	class := models.Class{
		ClassName: body.ClassName,
		ClassCode: body.ClassCode,
	}

	// fetch users based on provided IDs using for
	var users []*models.User
	for _, id := range body.UserIDs {
		var user models.User
		if result := initializers.DB.First(&user, id); result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found with ID: " + strconv.Itoa(int(id))})
			return
		}
		users = append(users, &user)
	}

	class.Users = users

	if result := initializers.DB.Create(&class); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"class": class})
}

func GetClassParticipants(c *gin.Context) {
	classID := c.Param("id")
	var class models.Class
	if result := initializers.DB.Preload("Users").First(&class, classID); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Class not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": class.Users})
}

func GetClassesUserEnrolled(c *gin.Context) {
	userID := c.Param("id")
	var user models.User
	if result := initializers.DB.Preload("Classes").First(&user, userID); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"classes": user.Classes})
}

func UpdateClassParticipants(c *gin.Context) {
	classID := c.Param("id")
	var body struct {
		UserIDs []uint `json:"user_ids" binding:"required"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var class models.Class
	if result := initializers.DB.First(&class, classID); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Class not found"})
		return
	}

	// fetch users based on provided IDs
	var users []*models.User
	for _, id := range body.UserIDs {
		var user models.User
		if result := initializers.DB.First(&user, id); result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found with ID: " + strconv.Itoa(int(id))})
			return
		}
		users = append(users, &user)
	}

	// use GORM association to replace many2many relations
	if err := initializers.DB.Model(&class).Association("Users").Replace(users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// reload class with users to return updated relation
	if result := initializers.DB.Preload("Users").First(&class, classID); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"class": class})
}

func DeleteParticipantFromClass(c *gin.Context) {
	classID := c.Param("id")
	userID := c.Param("user_id")

	var class models.Class
	if result := initializers.DB.First(&class, classID); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Class not found"})
		return
	}

	var user models.User
	if result := initializers.DB.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	// use GORM association to delete many2many relation
	if err := initializers.DB.Model(&class).Association("Users").Delete(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from class successfully"})
}

func DeleteUserFromAllClasses(c *gin.Context) {
	userID := c.Param("id")
	
	var user models.User
	if result := initializers.DB.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	// use GORM association to clear many2many relations
	if err := initializers.DB.Model(&user).Association("Classes").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User removed from all classes successfully"})
}