package controllers

import (
	"github.com/ElissaGunawan/moneywallbackend/initializers"
	"github.com/ElissaGunawan/moneywallbackend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CategoryCreate(c *gin.Context) {
	// get request data
	var body struct {
		CategoryName string
	}
	userID, err := getUserID(c)
	if err != nil {
		errorHandler(c, err)
		return
	}
	c.Bind(&body)

	// process create category
	category := models.Category{CategoryName: body.CategoryName, UserID: userID}
	err = initializers.DB.Transaction(func(tx *gorm.DB) error {
		result := initializers.DB.Create(&category)
		if result.Error != nil {
			return result.Error
		}
		return completeAchievement(tx, userID, achievementNewCategory)
	})
	if err != nil {
		errorHandler(c, err)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"category": category,
	})
}

func CategoryList(c *gin.Context) {
	// get request data
	userID, err := getUserID(c)
	if err != nil {
		errorHandler(c, err)
		return
	}
	page, perPage, err := getPaginationParam(c)
	if err != nil {
		errorHandler(c, err)
		return
	}

	// process list category
	var categories []models.Category
	result := initializers.DB.Limit(perPage).Offset((page-1)*perPage).Where("user_id = ?", userID).Find(&categories)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"categories": categories,
	})
}

func CategoryUpdate(c *gin.Context) {
	// get request data
	id := c.Param("id")
	var body struct {
		CategoryName string
	}
	c.Bind(&body)

	// process u[date] category
	var category models.Category
	result := initializers.DB.First(&category, id)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	category.CategoryName = body.CategoryName
	result = initializers.DB.Model(&category).Updates(category)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	//return response
	c.JSON(200, gin.H{
		"category": category,
	})
}

func CategoryDelete(c *gin.Context) {
	// get request data
	id := c.Param("id")

	// process delete category
	var category models.Category
	result := initializers.DB.First(&category, id)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	result = initializers.DB.Delete(&category)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.Status(200)
}
