package controllers

import (
	"github.com/ElissaGunawan/moneywallbackend/initializers"
	"github.com/ElissaGunawan/moneywallbackend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AccountCreate(c *gin.Context) {
	// get request data
	var body struct {
		AccountName string
		Amount      int
	}
	userID, err := getUserID(c)
	if err != nil {
		errorHandler(c, err)
		return
	}
	c.Bind(&body)

	// process create account
	var count int64
	result := initializers.DB.Model(&models.Account{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		errorHandler(c, result.Error)
		return
	}
	account := models.Account{AccountName: body.AccountName, Amount: body.Amount, FirstAmount: body.Amount, UserID: userID}
	err = initializers.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Create(&account)
		if result.Error != nil {
			return result.Error
		}
		if count >= 2 {
			txErr := unlockAvatar(tx, userID, avatarWomanLockedAchievement3Account)
			if txErr != nil {
				return txErr
			}
			errTx := completeAchievement(tx, userID, achievement3Account)
			if errTx != nil {
				return err
			}
		}
		return completeAchievement(tx, userID, achievementNewAccount)
	})
	if err != nil {
		errorHandler(c, err)
		return
	}

	// return success response
	c.JSON(200, gin.H{
		"account": account,
	})
}

func AccountList(c *gin.Context) {
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

	// process list account with pagination
	var accounts []models.Account
	result := initializers.DB.Limit(perPage).Offset((page-1)*perPage).Where("user_id = ?", userID).Find(&accounts)
	if result.Error != nil {
		c.JSON(200, gin.H{
			"accounts": accounts,
		})
		return
	}

	// return response
	c.JSON(200, gin.H{
		"accounts": accounts,
	})
}

func AccountUpdate(c *gin.Context) {
	// get request data
	id := c.Param("id")
	var body struct {
		AccountName string
	}
	c.Bind(&body)

	// process update account
	var account models.Account
	result := initializers.DB.First(&account, id)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	account.AccountName = body.AccountName
	result = initializers.DB.Model(&account).Updates(account)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"account": account,
	})
}

func AccountDelete(c *gin.Context) {
	// get request data
	id := c.Param("id")
	userID, err := getUserID(c)
	if err != nil {
		errorHandler(c, err)
		return
	}

	// process delete account
	var account models.Account
	result := initializers.DB.First(&account, id)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	result = initializers.DB.Delete(&models.Account{}, id, userID)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"account": account,
	})
}

func updateAccountAmount(tx *gorm.DB, accountID int, amount int) error {
	var account models.Account
	result := initializers.DB.First(&account, accountID)
	if result.Error != nil {
		return result.Error
	}
	account.Amount += amount
	result = tx.Model(&account).Updates(account)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
