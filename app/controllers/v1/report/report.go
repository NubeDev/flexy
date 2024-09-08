package reportController

import (
	"net/http"

	reportService "github.com/NubeDev/flexy/app/service/v1/report"
	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/gin-gonic/gin"
)

func Report(c *gin.Context) {
	appG := common.Gin{C: c}

	// Bind payload to struct
	var report reportService.ReportStruct
	if err := c.ShouldBindJSON(&report); err != nil {
		appG.Response(http.StatusBadRequest, code.InvalidParams, err.Error(), nil)
		return
	}

	// Validate bound struct parameters
	err, parameterErrorStr := common.CheckBindStructParameter(report, c)
	if err != nil {
		appG.Response(http.StatusBadRequest, code.InvalidParams, parameterErrorStr, nil)
		return
	}

	// Check if the report already exists
	var count = reportService.GetReportUserCountByPhoneAndActivityID(report.Phone, report.ActivityId)
	if count >= 1 {
		appG.Response(http.StatusBadRequest, code.InvalidParams, "Data already exists, please do not re-register!", nil)
		return
	}

	if report.ActivityId == 0 {
		report.ActivityId = 1
	}

	// Store information
	var reportResult = reportService.ReportInformation(report, c.ClientIP())

	m := make(map[string]interface{})
	m["uuid"] = reportResult.UUID
	m["name"] = report.Name
	// m["created_at"] = utils.TimeToDateTimesString(reportResult.CreatedAt)
	appG.Response(http.StatusOK, code.SUCCESS, "Information entered successfully!", m)
}
