package hostController

import (
	hostService "github.com/NubeDev/flexy/app/services/v1/host"
	"net/http"

	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/gin-gonic/gin"
)

var host = hostService.Get()

// CreateHost handles the creation of a new host.
func CreateHost(c *gin.Context) {
	g := common.Gin{C: c}
	var body *hostService.Fields
	if err := c.ShouldBindJSON(&body); err != nil {
		g.Response(http.StatusBadRequest, code.InvalidParams, err.Error(), nil)
		return
	}
	// Validate the parameters of the payload
	err, parameterErrorStr := common.CheckBindStructParameter(body, c)
	if err != nil {
		g.Response(http.StatusBadRequest, code.InvalidParams, parameterErrorStr, nil)
		return
	}
	// Call the service to create a new host
	resp, err := host.Create(body)
	if err != nil {
		g.Response(http.StatusInternalServerError, code.ERROR, err.Error(), nil)
		return
	}
	g.Response(http.StatusOK, code.SUCCESS, "Host created successfully", resp)
}

func GetHosts(c *gin.Context) {
	g := common.Gin{C: c}
	// Call the service to list all hosts
	hosts, err := host.GetHosts()
	if err != nil {
		g.Response(http.StatusInternalServerError, code.ERROR, err.Error(), nil)
		return
	}
	g.Response(http.StatusOK, code.SUCCESS, "success", hosts)
}

// GetHost retrieves a host by its UUID.
func GetHost(c *gin.Context) {
	g := common.Gin{C: c}
	uuid := c.Param("uuid")

	// Call the service to get the host by UUID
	host, err := host.GetHost(uuid)
	if err != nil {
		g.Response(http.StatusNotFound, code.ERROR, err.Error(), nil)
		return
	}
	g.Response(http.StatusOK, code.SUCCESS, "success", host)
}

// UpdateHost updates a host by its UUID.
func UpdateHost(c *gin.Context) {
	g := common.Gin{C: c}
	uuid := c.Param("uuid")

	var body *hostService.Fields
	if err := c.ShouldBindJSON(&body); err != nil {
		g.Response(http.StatusBadRequest, code.InvalidParams, err.Error(), nil)
		return
	}
	// Validate the payload
	err, parameterErrorStr := common.CheckBindStructParameter(body, c)
	if err != nil {
		g.Response(http.StatusBadRequest, code.InvalidParams, parameterErrorStr, nil)
		return
	}
	// Call the service to update the host
	host, err := host.Update(uuid, body)
	if err != nil {
		g.Response(http.StatusInternalServerError, code.ERROR, err.Error(), nil)
		return
	}
	g.Response(http.StatusOK, code.SUCCESS, "Host updated successfully", host)
}

// DeleteHost deletes a host by its UUID.
func DeleteHost(c *gin.Context) {
	g := common.Gin{C: c}
	uuid := c.Param("uuid")

	// Call the service to delete the host
	resp, err := host.Delete(uuid)
	if err != nil {
		g.Response(http.StatusInternalServerError, code.ERROR, err.Error(), nil)
		return
	}
	g.Response(http.StatusOK, code.SUCCESS, "Host deleted successfully", resp)
}
