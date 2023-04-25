package controllers

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/ElissaGunawan/moneywallbackend/initializers"
	"github.com/ElissaGunawan/moneywallbackend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Signup(c *gin.Context) {
	//Get the email/pass off req body
	var body struct {
		Email    string
		Name     string
		Password string
	}

	if c.Bind(&body) != nil {
		errorHandler(c, errors.New("failed to read body"))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		errorHandler(c, errors.New("failed to hash password"))
		return
	}

	user := models.User{
		Email:     body.Email,
		Name:      body.Name,
		Password:  string(hash),
		AvatarURL: avatarDefaultURL,
	}
	err = initializers.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Create(&user)
		if result.Error != nil {
			return result.Error
		}
		err = initAchievement(tx, int(user.ID))
		if err != nil {
			return err
		}
		err = initQuest(tx, int(user.ID))
		if err != nil {
			return err
		}
		return initAvatar(tx, int(user.ID))
	})
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func Login(c *gin.Context) {
	//Get the email and pass off req body
	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}

	//Look up requested user
	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	//Compare
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password",
		})
		return
	}
	//Generate a jwt token

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid to create token",
		})
		return
	}
	//Send it back
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}

func Validate(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"message": user,
	})
}

func Profile(c *gin.Context) {
	// get request data
	userID, err := getUserID(c)
	if err != nil {
		errorHandler(c, err)
		return
	}

	// process
	var user models.User
	result := initializers.DB.First(&user, userID)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"user": user,
	})
}

func Leaderboard(c *gin.Context) {
	// get request data
	page, perPage, err := getPaginationParam(c)
	if err != nil {
		errorHandler(c, err)
		return
	}

	// process list users with pagination
	var users []models.User
	result := initializers.DB.Limit(perPage).Offset((page - 1) * perPage).Order("exp desc, id asc").Find(&users)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"users": users,
	})
}

func ProfileUpdate(c *gin.Context) {
	// get request data
	userID, err := getUserID(c)
	if err != nil {
		errorHandler(c, err)
		return
	}
	var body struct {
		Name      string
		AvatarURL string
	}
	c.Bind(&body)

	// process update account
	var user models.User
	result := initializers.DB.First(&user, userID)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}
	if body.Name != "" {
		user.Name = body.Name
	}
	if body.AvatarURL != "" {
		user.AvatarURL = body.AvatarURL
	}

	err = initializers.DB.Transaction(func(tx *gorm.DB) error {
		result = initializers.DB.Model(&user).Updates(user)
		if result.Error != nil {
			return result.Error
		}
		txErr := unlockAvatar(tx, userID, avatarManLockedAchievementEditProfile)
		if txErr != nil {
			return txErr
		}
		return completeAchievement(tx, userID, achievementEditProfile)
	})
	if err != nil {
		errorHandler(c, err)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"user": user,
	})
}
