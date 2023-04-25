package controllers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ElissaGunawan/moneywallbackend/initializers"
	"github.com/ElissaGunawan/moneywallbackend/models"
	"github.com/ElissaGunawan/moneywallbackend/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	dateFormat = "02/01/2006"
)

type IncomeResponse struct {
	ID          uint
	UserID      int
	Date        string
	AccountID   int
	AccountName string
	IncomeName  string
	Amount      int
}

type IncomeDashboardResponse struct {
	AccountID   int
	AccountName string
	Amount      int
}

func IncomeCreate(c *gin.Context) {
	// get request data
	var body struct {
		Date       string
		IncomeName string
		AccountID  int
		Amount     int
	}
	c.Bind(&body)
	userID, err := getUserID(c)
	if err != nil {
		errorHandler(c, err)
		return
	}
	dateTime, err := time.Parse(dateFormat, body.Date)
	if err != nil {
		errorHandler(c, err)
		return
	}

	// process create income
	var count int64
	result := initializers.DB.Model(&models.Income{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		errorHandler(c, result.Error)
		return
	}
	log.Printf("count = %d", count)
	income := models.Income{Date: dateTime, IncomeName: body.IncomeName, Amount: body.Amount, UserID: userID, AccountID: body.AccountID}
	err = initializers.DB.Transaction(func(tx *gorm.DB) error {
		result := initializers.DB.Create(&income)
		if result.Error != nil {
			return result.Error
		}
		txErr := updateAccountAmount(tx, body.AccountID, body.Amount)
		if txErr != nil {
			return txErr
		}
		if count >= 4 {
			txErr = unlockAvatar(tx, userID, avatarManLockedAchievement5Income)
			if txErr != nil {
				return txErr
			}
			txErr = completeAchievement(tx, userID, achievement5Income)
			if txErr != nil {
				return txErr
			}
		}
		return completeQuest(tx, userID, questFirstIncome)
	})
	if err != nil {
		errorHandler(c, err)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"income": income,
	})
}

func IncomeList(c *gin.Context) {
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
	startDate, endDate, err := getStartEndDateParam(c)
	if err != nil {
		errorHandler(c, err)
		return
	}

	// process list incomes
	var incomes []models.Income
	var accounts []models.Account
	result := initializers.DB.
		Where("user_id = ?", userID).
		Where("date >= ? AND date <= ?", startDate, endDate).
		Limit(perPage).Offset((page - 1) * perPage).
		Find(&incomes)
	if result.Error != nil {
		errorHandler(c, err)
		return
	}
	if len(incomes) == 0 {
		c.JSON(200, gin.H{
			"incomes": convertArrayToIncomeResponse(incomes, accounts),
		})
		return
	}
	accountIDs := make([]string, 0)
	for _, income := range incomes {
		accountIDs = append(accountIDs, strconv.Itoa(income.AccountID))
	}
	result = initializers.DB.Where(fmt.Sprintf("id in (%s)", strings.Join(util.UniqueString(accountIDs), ","))).Find(&accounts)
	if result.Error != nil {
		errorHandler(c, err)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"incomes": convertArrayToIncomeResponse(incomes, accounts),
	})
}

func convertToIncomeDashboard(incomes []models.IncomeDashboard, accounts []models.Account) []IncomeDashboardResponse {
	result := make([]IncomeDashboardResponse, 0)
	for _, v := range incomes {
		result = append(result, convertToIncomeDashboardResponse(v, accounts))
	}
	return result
}

func convertArrayToIncomeResponse(incomes []models.Income, accounts []models.Account) []IncomeResponse {
	result := make([]IncomeResponse, 0)
	for _, v := range incomes {
		result = append(result, convertToIncomeResponse(v, accounts))
	}
	return result
}

func convertToIncomeResponse(income models.Income, accounts []models.Account) IncomeResponse {
	var accountName string
	for _, account := range accounts {
		if account.ID == uint(income.AccountID) {
			accountName = account.AccountName
		}
	}
	return IncomeResponse{
		ID:          income.ID,
		UserID:      income.UserID,
		Date:        income.Date.Format(dateFormat),
		AccountID:   income.AccountID,
		AccountName: accountName,
		IncomeName:  income.IncomeName,
		Amount:      income.Amount,
	}
}

func convertToIncomeDashboardResponse(income models.IncomeDashboard, accounts []models.Account) IncomeDashboardResponse {
	var accountName string
	for _, account := range accounts {
		if account.ID == uint(income.AccountID) {
			accountName = account.AccountName
		}
	}
	return IncomeDashboardResponse{
		AccountID:   income.AccountID,
		AccountName: accountName,
		Amount:      income.Amount,
	}
}

func IncomeUpdate(c *gin.Context) {
	// get request data
	id := c.Param("id")
	var body struct {
		Date       string
		IncomeName string
		AccountID  int
		Amount     int
	}
	c.Bind(&body)
	date, err := time.Parse(dateFormat, body.Date)
	if err != nil {
		errorHandler(c, err)
		return
	}

	// process update income
	var income models.Income
	var diffAmount, oldAmount, oldAccountID int
	var sameAccount bool
	result := initializers.DB.First(&income, id)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	if body.IncomeName != "" {
		income.IncomeName = body.IncomeName
	}
	if body.AccountID != 0 {
		if income.AccountID == body.AccountID {
			sameAccount = true
		}
		oldAccountID = income.AccountID
		income.AccountID = body.AccountID
	}
	if body.Amount != 0 {
		oldAmount = body.Amount
		diffAmount = body.Amount - income.Amount
		income.Amount = body.Amount
	}
	income.Date = date

	err = initializers.DB.Transaction(func(tx *gorm.DB) error {
		result = initializers.DB.Model(&income).Updates(income)
		if result.Error != nil {
			return result.Error
		}
		if sameAccount {
			return updateAccountAmount(tx, income.AccountID, diffAmount)
		}
		err := updateAccountAmount(tx, oldAccountID, -oldAmount)
		if err != nil {
			return err
		}
		return updateAccountAmount(tx, income.AccountID, body.Amount)
	})
	if err != nil {
		errorHandler(c, err)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"income": income,
	})
}

func IncomeDelete(c *gin.Context) {
	// get request data
	id := c.Param("id")

	// process delete income
	var income models.Income
	result := initializers.DB.First(&income, id)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		result := initializers.DB.Delete(&income)
		if result.Error != nil {
			return result.Error
		}
		return updateAccountAmount(tx, income.AccountID, -income.Amount)
	})
	if err != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"income": income,
	})
}

func IncomeDashboard(c *gin.Context) {
	// get request data
	userID, err := getUserID(c)
	if err != nil {
		errorHandler(c, err)
		return
	}
	startDate, endDate, err := getStartEndDateParam(c)
	if err != nil {
		errorHandler(c, err)
		return
	}

	var incomeDashboard []models.IncomeDashboard
	var accounts []models.Account
	result := initializers.DB.Table("incomes").Select("account_id, sum(amount) as amount").
		Where("date >= ? AND date <= ? AND user_id = ? AND deleted_at IS NULL", startDate, endDate, userID).
		Group("account_id").Find(&incomeDashboard)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	if len(incomeDashboard) == 0 {
		c.JSON(200, gin.H{
			"incomes": convertToIncomeDashboard(incomeDashboard, accounts),
		})
		return
	}
	accountIDs := make([]string, 0)
	for _, income := range incomeDashboard {
		accountIDs = append(accountIDs, strconv.Itoa(income.AccountID))
	}
	result = initializers.DB.Where(fmt.Sprintf("id in (%s)", strings.Join(util.UniqueString(accountIDs), ","))).Find(&accounts)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"incomes": convertToIncomeDashboard(incomeDashboard, accounts),
	})
}
