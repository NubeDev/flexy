package utils

import (
	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/helpers/com"
	"github.com/gin-gonic/gin"
)

type PageResult struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	//Page     int         `json:"page"`
	PageSize    int `json:"pageSize"`
	CurrentPage int `json:"currentPage"`
}

type Page struct {
	P uint `json:"p" form:"p" validate:"numeric" default:"1"`
	N uint `json:"n" form:"n" validate:"numeric" default:"15"`
}

// GetPage retrieves page parameters from the query string
func GetPage(c *gin.Context) (error, string, int, int) {
	currentPage := 0
	// Bind query parameters to a struct
	var p Page
	if err := c.ShouldBindQuery(&p); err != nil {
		return err, "Failed to bind parameters, please check the type of the parameters!", 0, 0
	}
	// Validate bound struct parameters
	err, parameterErrorStr := common.CheckBindStructParameter(p, c)
	if err != nil {
		return err, parameterErrorStr, 0, 0
	}
	// Get the page and limit from query parameters, defaulting to "0" and "15" respectively
	page := com.StrTo(c.DefaultQuery("p", "0")).MustInt()
	limit := com.StrTo(c.DefaultQuery("n", "15")).MustInt()

	// Calculate the starting point for the current page
	if page > 0 {
		currentPage = (page - 1) * limit
	}
	return nil, "", currentPage, limit
}
