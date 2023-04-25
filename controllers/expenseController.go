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

type ExpenseResponse struct {
	ID           uint
	UserID       int
	Date         string
	AccountID    int
	CategoryID   int
	CategoryName string
	AccountName  string
	Amount       int
	ExpenseName  string
}

type ExpenseDashboardResponse struct {
	CategoryID   int
	CategoryName string
	Amount       int
}

func ExpenseCreate(c *gin.Context) {
	// get request data
	var body struct {
		Date        string
		AccountID   int
		CategoryID  int
		Amount      int
		ExpenseName string
	}
	c.Bind(&body)
	userID, err := getUserID(c)
	if err != nil {
		errorHandler(c, err)
		return
	}
	date, err := time.Parse(dateFormat, body.Date)
	if err != nil {
		errorHandler(c, err)
		return
	}

	// process create expense
	var count int64
	result := initializers.DB.Model(&models.Expense{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		errorHandler(c, result.Error)
		return
	}
	log.Printf("count = %d", count)
	expense := models.Expense{Date: date, AccountID: body.AccountID, CategoryID: body.CategoryID, Amount: body.Amount, ExpenseName: body.ExpenseName, UserID: userID}
	err = initializers.DB.Transaction(func(tx *gorm.DB) error {
		result := initializers.DB.Create(&expense)
		if result.Error != nil {
			return result.Error
		}
		txErr := updateAccountAmount(tx, body.AccountID, -body.Amount)
		if txErr != nil {
			return txErr
		}
		if count >= 4 {
			txErr = unlockAvatar(tx, userID, avatarWomanLockedAchievement5Expense)
			if txErr != nil {
				return txErr
			}
			txErr = completeAchievement(tx, userID, achievement5Expense)
			if txErr != nil {
				return txErr
			}
		}
		return completeQuest(tx, userID, questFirstExpense)
	})
	if err != nil {
		errorHandler(c, err)
		return
	}

	//return response
	c.JSON(200, gin.H{
		"expense": expense,
	})
}

func ExpenseList(c *gin.Context) {
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

	// process expense list
	var expenses []models.Expense
	var categories []models.Category
	var accounts []models.Account
	result := initializers.DB.
		Where("user_id = ?", userID).
		Where("date >= ? AND date <= ?", startDate, endDate).
		Limit(perPage).Offset((page - 1) * perPage).
		Find(&expenses)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	if len(expenses) == 0 {
		c.JSON(200, gin.H{
			"expenses": convertArrayToExpenseResponse(expenses, categories, accounts),
		})
		return
	}
	categoryIDs := make([]string, 0)
	for _, expense := range expenses {
		categoryIDs = append(categoryIDs, strconv.Itoa(expense.CategoryID))
	}
	result = initializers.DB.
		Where(fmt.Sprintf("id in (%s)", strings.Join(util.UniqueString(categoryIDs), ","))).
		Find(&categories)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	accountIDs := make([]string, 0)
	for _, expense := range expenses {
		accountIDs = append(accountIDs, strconv.Itoa(expense.AccountID))
	}
	result = initializers.DB.Where(fmt.Sprintf("id in (%s)", strings.Join(util.UniqueString(accountIDs), ","))).Find(&accounts)
	if result.Error != nil {
		errorHandler(c, err)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"expenses": convertArrayToExpenseResponse(expenses, categories, accounts),
	})
}
func ExpenseUpdate(c *gin.Context) {
	// get request data
	id := c.Param("id")
	var body struct {
		Date        string
		AccountID   int
		CategoryID  int
		Amount      int
		ExpenseName string
	}
	c.Bind(&body)
	date, err := time.Parse(dateFormat, body.Date)
	if err != nil {
		errorHandler(c, err)
		return
	}

	// process update expense
	var expense models.Expense
	var diffAmount, oldAmount, oldAccountID int
	var sameAccount bool
	result := initializers.DB.First(&expense, id)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	if body.ExpenseName != "" {
		expense.ExpenseName = body.ExpenseName
	}
	if body.CategoryID != 0 {
		expense.CategoryID = body.CategoryID
	}
	if body.AccountID != 0 {
		if expense.AccountID == body.AccountID {
			sameAccount = true
		}
		oldAccountID = expense.AccountID
		expense.AccountID = body.AccountID
	}
	if body.Amount != 0 {
		oldAmount = body.Amount
		diffAmount = body.Amount - expense.Amount
		expense.Amount = body.Amount
	}
	expense.Date = date

	err = initializers.DB.Transaction(func(tx *gorm.DB) error {
		result = initializers.DB.Model(&expense).Updates(expense)
		if result.Error != nil {
			return result.Error
		}
		if sameAccount {
			return updateAccountAmount(tx, expense.AccountID, -diffAmount)
		}
		err := updateAccountAmount(tx, oldAccountID, oldAmount)
		if err != nil {
			return err
		}
		return updateAccountAmount(tx, expense.AccountID, -body.Amount)
	})
	if err != nil {
		errorHandler(c, err)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"expense": expense,
	})
}

func ExpenseDelete(c *gin.Context) {
	// get request data
	id := c.Param("id")

	// process delete expense
	var expense models.Expense
	result := initializers.DB.First(&expense, id)
	if result.Error != nil {
		c.JSON(200, gin.H{
			"expense": expense,
		})
		return
	}
	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		result := initializers.DB.Delete(&expense)
		if result.Error != nil {
			return result.Error
		}
		return updateAccountAmount(tx, expense.AccountID, expense.Amount)
	})
	if err != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"expense": expense,
	})
}

func convertArrayToExpenseResponse(expenses []models.Expense, categories []models.Category, accounts []models.Account) []ExpenseResponse {
	result := make([]ExpenseResponse, 0)
	for _, v := range expenses {
		result = append(result, convertToExpenseResponse(v, categories, accounts))
	}
	return result
}

func convertToExpenseResponse(expense models.Expense, categories []models.Category, accounts []models.Account) ExpenseResponse {
	var categoryName, accountName string
	for _, category := range categories {
		if category.ID == uint(expense.CategoryID) {
			categoryName = category.CategoryName
		}
	}
	for _, account := range accounts {
		if account.ID == uint(expense.AccountID) {
			accountName = account.AccountName
		}
	}

	return ExpenseResponse{
		ID:           expense.ID,
		UserID:       expense.UserID,
		Date:         expense.Date.Format(dateFormat),
		AccountID:    expense.AccountID,
		CategoryID:   expense.CategoryID,
		Amount:       expense.Amount,
		ExpenseName:  expense.ExpenseName,
		CategoryName: categoryName,
		AccountName:  accountName,
	}
}
func ExpenseDashboard(c *gin.Context) {
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

	var expenseDashboard []models.ExpenseDashboard
	var categories []models.Category
	result := initializers.DB.Table("expenses").Select("category_id, sum(amount) as amount").
		Where("date >= ? AND date <= ? AND user_id = ? AND deleted_at IS NULL", startDate, endDate, userID).
		Group("category_id").Find(&expenseDashboard)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	if len(expenseDashboard) == 0 {
		c.JSON(200, gin.H{
			"expenses": convertToExpenseDashboard(expenseDashboard, categories),
		})
		return
	}
	categoryIDs := make([]string, 0)
	for _, expense := range expenseDashboard {
		categoryIDs = append(categoryIDs, strconv.Itoa(expense.CategoryID))
	}
	result = initializers.DB.Where(fmt.Sprintf("id in (%s)", strings.Join(util.UniqueString(categoryIDs), ","))).Find(&categories)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"expenses": convertToExpenseDashboard(expenseDashboard, categories),
	})
}
func convertToExpenseDashboard(expenses []models.ExpenseDashboard, categories []models.Category) []ExpenseDashboardResponse {
	result := make([]ExpenseDashboardResponse, 0)
	for _, v := range expenses {
		result = append(result, convertToExpenseDashboardResponse(v, categories))
	}
	return result
}

func convertToExpenseDashboardResponse(expense models.ExpenseDashboard, categories []models.Category) ExpenseDashboardResponse {
	var categoryName string
	for _, category := range categories {
		if category.ID == uint(expense.CategoryID) {
			categoryName = category.CategoryName
		}
	}
	return ExpenseDashboardResponse{
		CategoryID:   expense.CategoryID,
		CategoryName: categoryName,
		Amount:       expense.Amount,
	}
}
