package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ElissaGunawan/moneywallbackend/models"
	"github.com/gin-gonic/gin"
)

// getUserID will get user_id from context derived from header Authorization
func getUserID(c *gin.Context) (int, error) {
	obj, exist := c.Get("user")
	if !exist {
		return 0, errors.New("user not found")
	}
	user, ok := obj.(models.User)
	if !ok {
		return 0, errors.New("user not found")
	}
	return int(user.ID), nil
}

// getPaginationParam get param page and per_page
func getPaginationParam(c *gin.Context) (int, int, error) {
	pageString, ok := c.GetQuery("page")
	if !ok {
		return 0, 0, errors.New("param page not found")
	}
	page, err := strconv.Atoi(pageString)
	if err != nil {
		return 0, 0, errors.New("param page not found")
	}

	perPageString, ok := c.GetQuery("per_page")
	if !ok {
		return 0, 0, errors.New("param per_page not found")
	}
	perPage, err := strconv.Atoi(perPageString)
	if err != nil {
		return 0, 0, errors.New("param per_page not found")
	}
	return page, perPage, nil
}

// getStartEndDateParam from param
func getStartEndDateParam(c *gin.Context) (*time.Time, *time.Time, error) {
	startDateString, ok := c.GetQuery("start_date")
	if !ok {
		return nil, nil, errors.New("param start_date not found")
	}
	endDateString, ok := c.GetQuery("end_date")
	if !ok {
		return nil, nil, errors.New("param end_date not found")
	}
	return parseStartEndDateTimeParam(startDateString, endDateString)
}

// getStartEndDateTimeParam parse time
func parseStartEndDateTimeParam(startDate, endDate string) (*time.Time, *time.Time, error) {
	startDateTime, err := time.Parse(dateFormat, startDate)
	if err != nil {
		return nil, nil, err
	}
	endDateTime, err := time.Parse(dateFormat, endDate)
	if err != nil {
		return nil, nil, err
	}
	return &startDateTime, &endDateTime, nil
}

func errorHandler(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": fmt.Sprintf("%v", err),
	})
}
