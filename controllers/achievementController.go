package controllers

import (
	"time"

	"github.com/ElissaGunawan/moneywallbackend/initializers"
	"github.com/ElissaGunawan/moneywallbackend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	achievementNewCategory      = "new_category"
	achievementNewAccount       = "new_account"
	achievement5Expense         = "5_expense"
	achievement5Income          = "5_income"
	achievement3Account         = "3_account"
	achievementEditProfile      = "edit_profile"
	achievementCompleteAllQuest = "complete_all_quest"
	achievementReduceExpense    = "reduce_expense"
	achievementBeginnersLuck    = "beginners_luck"
)

var initAchievements = []models.Achievement{
	{
		AchievementCode: achievementNewCategory,
		AchievementName: "Membuat kategori baru",
		Exp:             50,
		Reward:          "Get 50 exp",
	},
	{
		AchievementCode: achievementNewAccount,
		AchievementName: "Membuat akun baru",
		Exp:             50,
		Reward:          "Get 50 exp",
	},
	{
		AchievementCode: achievement5Expense,
		AchievementName: "Membuat 5 pengeluaran baru berturut-turut",
		Reward:          "Unlock a new avatar",
	},
	{
		AchievementCode: achievement5Income,
		AchievementName: "Membuat 5 pemasukan baru berturut-turut",
		Reward:          "Unlock a new avatar",
	},
	{
		AchievementCode: achievement3Account,
		AchievementName: "Membuat 3 akun baru berturut-turut",
		Reward:          "Unlock a new avatar",
	},
	{
		AchievementCode: achievementEditProfile,
		AchievementName: "Mengubah avatar profile",
		Reward:          "Unlock a new avatar",
	},
	{
		AchievementCode: achievementCompleteAllQuest,
		AchievementName: "Selesaikan semua quest",
		Reward:          "Unlock a new avatar",
	},
}

func initAchievement(tx *gorm.DB, userID int) error {
	for _, v := range initAchievements {
		result := tx.Create(&models.Achievement{
			UserID:          userID,
			AchievementCode: v.AchievementCode,
			AchievementName: v.AchievementName,
			Exp:             v.Exp,
			Reward:          v.Reward,
		})
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func completeAchievement(tx *gorm.DB, userID int, achievementCode string) error {
	var achievement models.Achievement
	var user models.User
	now := time.Now()
	result := initializers.DB.
		Where("user_id = ?", userID).
		Where("achievement_code = ?", achievementCode).
		First(&achievement)
	if result.Error != nil {
		return result.Error
	}
	if achievement.CompletedAt != nil { // achievement has already been completed
		return nil
	}
	achievement.CompletedAt = &now
	result = tx.Model(&achievement).Where("id = ?", achievement.ID).Updates(achievement)
	if result.Error != nil {
		return result.Error
	}
	result = initializers.DB.First(&user, userID)
	if result.Error != nil {
		return result.Error
	}
	user.Exp += achievement.Exp
	result = tx.Model(&user).Updates(user)
	return result.Error
}

func AchievementList(c *gin.Context) {
	// get reachievement data
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
	var achievements []models.Achievement
	result := initializers.DB.Limit(perPage).Offset((page-1)*perPage).Where("user_id = ?", userID).Find(&achievements)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"achievements": achievements,
	})
}
