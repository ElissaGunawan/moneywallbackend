package controllers

import (
	"fmt"
	"log"
	"time"

	"github.com/ElissaGunawan/moneywallbackend/initializers"
	"github.com/ElissaGunawan/moneywallbackend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	questFirstExpense = "first_expense"
	questFirstIncome  = "first_income"
	questSecondIncome = "second_income"
	questNewAccount   = "new_account"
)

var initQuests = []models.Quest{
	{
		QuestCode: questFirstExpense,
		QuestName: "Memasukkan pengeluaran pertama hari ini",
		Exp:       20,
		Reward:    "Get 20 exp",
	},
	{
		QuestCode: questFirstIncome,
		QuestName: "Memasukkan pemasukkan pertama hari ini",
		Exp:       20,
		Reward:    "Get 20 exp",
	},
}

func initQuest(tx *gorm.DB, userID int) error {
	for _, v := range initQuests {
		result := tx.Create(&models.Quest{
			UserID:    userID,
			QuestCode: v.QuestCode,
			QuestName: v.QuestName,
			Exp:       v.Exp,
			Reward:    v.Reward,
		})
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func completeQuest(tx *gorm.DB, userID int, questCode string) error {
	var quest models.Quest
	var user models.User
	now := time.Now()
	result := initializers.DB.
		Where("user_id = ?", userID).
		Where("quest_code = ?", questCode).
		First(&quest)
	if result.Error != nil {
		return result.Error
	}
	if quest.CompletedAt != nil { // quest has already been completed
		return nil
	}
	quest.CompletedAt = &now
	result = tx.Model(&quest).Where("id = ?", quest.ID).Updates(quest)
	if result.Error != nil {
		return result.Error
	}
	// process list account with pagination
	var quests []models.Quest
	result = initializers.DB.Where("user_id = ?", userID).Find(&quests)
	if result.Error != nil {
		return result.Error
	}
	isAllQuestCompleted := true
	for _, q := range quests {
		if q.CompletedAt == nil && q.QuestCode != questCode {
			isAllQuestCompleted = false
		}
	}
	if isAllQuestCompleted {
		err := completeAchievement(tx, userID, achievementCompleteAllQuest)
		if err != nil {
			return err
		}
		err = unlockAvatar(tx, userID, avatarManLockedCompletedDailyQuest)
		if err != nil {
			return err
		}
	}

	result = initializers.DB.First(&user, userID)
	if result.Error != nil {
		return result.Error
	}
	user.Exp += quest.Exp
	result = tx.Model(&user).Updates(user)
	return result.Error
}

func RefreshQuestHandler(c *gin.Context) {
	err := RefreshQuest()
	if err != nil {
		errorHandler(c, err)
		return
	}
	c.Status(200)
}

func RefreshQuest() error {
	result := initializers.DB.Model(models.Quest{}).Where("completed_at IS NOT NULL").UpdateColumn("CompletedAt", nil)
	if result.Error != nil {
		return result.Error
	}
	var users []models.User
	result = initializers.DB.Find(&users)
	if result.Error != nil {
		return result.Error
	}
	for _, user := range users {
		var firstExpense models.Expense
		result := initializers.DB.
			Where("user_id = ?", user.ID).
			Order("date asc, id asc").
			First(&firstExpense)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			return result.Error
		}
		if time.Now().AddDate(0, 0, -7).Before(firstExpense.Date) {
			log.Printf("user %d first expense is at %s, so user is not eligible for reduce expense achievement", user.ID, firstExpense.Date.Format(dateFormat))
			continue
		}
		var expenseSum float32
		result = initializers.DB.Table("expenses").Select("COALESCE(SUM(amount),0) as amount").
			Where("date >= ? AND date <= ? AND user_id = ? AND deleted_at IS NULL", time.Now().AddDate(0, 0, -7), time.Now(), user.ID).
			Find(&expenseSum)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			return result.Error
		}
		if expenseSum > 0 {
			var achievement models.Achievement
			result := initializers.DB.
				Where("user_id = ?", user.ID).
				Where("achievement_code = ?", achievementReduceExpense).
				First(&achievement)
			if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
				return result.Error
			}
			if achievement.ID >= 1 {
				var todayExpenseSum float32
				result = initializers.DB.Table("expenses").Select("COALESCE(SUM(amount),0) as amount").
					Where("date >= ? AND date <= ? AND user_id = ? AND deleted_at IS NULL", time.Now().AddDate(0, 0, -1), time.Now(), user.ID).
					Find(&expenseSum)
				if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
					return result.Error
				}
				if todayExpenseSum < achievement.ReduceExpense {
					err := initializers.DB.Transaction(func(tx *gorm.DB) error {
						return completeAchievement(tx, achievement.UserID, achievementReduceExpense)
					})
					if err != nil {
						return err
					}
				}
				continue
			}
			result = initializers.DB.Create(&models.Achievement{
				UserID:          int(user.ID),
				AchievementCode: achievementReduceExpense,
				AchievementName: fmt.Sprintf("Kurangi pengeluaran harian sebesar Rp %.2f (5 persen pengeluaran harian)", expenseSum/7*0.05),
				ReduceExpense:   expenseSum / 7 * 0.05,
				Exp:             50,
				Reward:          "Get 50 exp",
			})
			if result.Error != nil {
				return result.Error
			}
		}
	}

	return nil
}

func QuestList(c *gin.Context) {
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
	var quests []models.Quest
	result := initializers.DB.Limit(perPage).Offset((page-1)*perPage).Where("user_id = ?", userID).Find(&quests)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"quests": quests,
	})
}
