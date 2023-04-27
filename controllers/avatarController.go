package controllers

import (
	"time"

	"github.com/ElissaGunawan/moneywallbackend/initializers"
	"github.com/ElissaGunawan/moneywallbackend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	avatarDefault                             = "avatar_default"
	avatarWomanLockedAchievement5Expense      = "avatar_woman"
	avatarMan                                 = "avatar_man"
	avatarManBrownHair                        = "avatar_man_brown_hair"
	avatarWomanCurlyHair                      = "avatar_woman_curly_hair"
	avatarManBlackHair                        = "avatar_man_black_hair"
	avatarManLockedAchievement5Income         = "avatar_man_locked_achievement"
	avatarWomanLockedAchievement3Account      = "avatar_woman_locked_achievement"
	avatarManLockedCompletedDailyQuest        = "avatar_man_locked_daily_quest"
	avatarManLockedAchievementEditProfile     = "avatar_man_locked_edit_profile"
	avatarPandaLockedAchievementBeginnersLuck = "avatar_panda"
	avatarDefaultURL                          = "https://ik.imagekit.io/cq2exqppw/man__4_.png"
)

var initAvatars = []models.Avatar{
	{
		AvatarCode: avatarDefault,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/man__4_.png",
	},
	{
		AvatarCode: avatarWomanLockedAchievement5Expense,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/woman.png",
	},
	{
		AvatarCode: avatarMan,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/profile__1_.png",
	},
	{
		AvatarCode: avatarManBrownHair,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/man__1_.png",
	},
	{
		AvatarCode: avatarWomanCurlyHair,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/woman__2_.png",
	},
	{
		AvatarCode: avatarManBlackHair,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/man__2_.png",
	},
	{
		AvatarCode: avatarManLockedAchievement5Income,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/man__3_.png",
	},
	{
		AvatarCode: avatarWomanLockedAchievement3Account,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/woman__1_.png",
	},
	{
		AvatarCode: avatarManLockedCompletedDailyQuest,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/man.png",
	},
	{
		AvatarCode: avatarManLockedAchievementEditProfile,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/man__5_.png",
	},
	{
		AvatarCode: avatarPandaLockedAchievementBeginnersLuck,
		AvatarURL:  "https://ik.imagekit.io/cq2exqppw/panda.png",
	},
}

func initAvatar(tx *gorm.DB, userID int) error {
	now := time.Now()
	for _, v := range initAvatars {
		avatar := models.Avatar{
			UserID:     userID,
			AvatarCode: v.AvatarCode,
			AvatarURL:  v.AvatarURL,
		}
		if v.AvatarCode == avatarDefault ||
			v.AvatarCode == avatarMan ||
			v.AvatarCode == avatarManBrownHair ||
			v.AvatarCode == avatarWomanCurlyHair ||
			v.AvatarCode == avatarManBlackHair {
			avatar.UnlockedAt = &now
		}
		result := tx.Create(&avatar)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func unlockAvatar(tx *gorm.DB, userID int, avatarCode string) error {
	var avatar models.Avatar
	now := time.Now()
	result := initializers.DB.
		Where("user_id = ?", userID).
		Where("avatar_code = ?", avatarCode).
		First(&avatar)
	if result.Error != nil {
		return result.Error
	}
	if avatar.UnlockedAt != nil {
		return nil
	}
	avatar.UnlockedAt = &now
	result = tx.Model(&avatar).Where("id = ?", avatar.ID).Updates(avatar)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func AvatarList(c *gin.Context) {
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
	var avatars []models.Avatar
	result := initializers.DB.Limit(perPage).Offset((page-1)*perPage).Where("user_id = ?", userID).Find(&avatars)
	if result.Error != nil {
		errorHandler(c, result.Error)
		return
	}

	// return response
	c.JSON(200, gin.H{
		"avatars": avatars,
	})
}
