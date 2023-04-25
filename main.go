package main

import (
	"log"
	"time"

	"github.com/ElissaGunawan/moneywallbackend/controllers"
	"github.com/ElissaGunawan/moneywallbackend/initializers"
	"github.com/ElissaGunawan/moneywallbackend/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	initializers.SyncDatabase()
}

func main() {
	initCron()
	r := gin.Default()

	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)
	r.GET("/validate", middleware.RequireAuth, controllers.Validate)

	r.POST("/accounts", middleware.RequireAuth, controllers.AccountCreate)
	r.GET("/accounts", middleware.RequireAuth, controllers.AccountList)
	r.POST("/accounts/:id", middleware.RequireAuth, controllers.AccountUpdate)
	r.DELETE("/accounts/:id", middleware.RequireAuth, controllers.AccountDelete)

	r.POST("/categories", middleware.RequireAuth, controllers.CategoryCreate)
	r.GET("/categories", middleware.RequireAuth, controllers.CategoryList)
	r.POST("/categories/:id", middleware.RequireAuth, controllers.CategoryUpdate)
	r.DELETE("/categories/:id", middleware.RequireAuth, controllers.CategoryDelete)

	r.POST("/incomes", middleware.RequireAuth, controllers.IncomeCreate)
	r.GET("/incomes", middleware.RequireAuth, controllers.IncomeList)
	r.POST("/incomes/:id", middleware.RequireAuth, controllers.IncomeUpdate)
	r.DELETE("/incomes/:id", middleware.RequireAuth, controllers.IncomeDelete)
	r.GET("/incomedashboard", middleware.RequireAuth, controllers.IncomeDashboard)

	r.POST("/expenses", middleware.RequireAuth, controllers.ExpenseCreate)
	r.GET("/expenses", middleware.RequireAuth, controllers.ExpenseList)
	r.POST("/expenses/:id", middleware.RequireAuth, controllers.ExpenseUpdate)
	r.DELETE("/expenses/:id", middleware.RequireAuth, controllers.ExpenseDelete)
	r.GET("/expensedashboard", middleware.RequireAuth, controllers.ExpenseDashboard)

	r.GET("/quests", middleware.RequireAuth, controllers.QuestList)
	r.GET("/achievements", middleware.RequireAuth, controllers.AchievementList)
	r.GET("/leaderboard", controllers.Leaderboard)
	r.GET("/profile", middleware.RequireAuth, controllers.Profile)
	r.PUT("/profile", middleware.RequireAuth, controllers.ProfileUpdate)

	r.GET("/avatars", middleware.RequireAuth, controllers.AvatarList)
	r.GET("/refresh-quest", middleware.RequireAuth, controllers.RefreshQuestHandler)
	r.Run()
}

func initCron() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At("00:00").Do(func() {
		err := controllers.RefreshQuest()
		if err != nil {
			log.Printf("daily refresh quest failed, err=%v", err)
		} else {
			log.Printf("daily refresh quest success")
		}
	})
	s.StartAsync()
}
