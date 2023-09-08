package util

import (
	"log"
	"mofe64/playlistGen/data/responses"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var ErrorLog = log.New(os.Stdout, "error  --> ", log.LstdFlags)
var InfoLog = log.New(os.Stdout, "info --> ", log.LstdFlags)

func GenerateJSONResponse(c *gin.Context, statusCode int, message string, data map[string]interface{}) {
	c.JSON(statusCode, responses.APIResponse{
		Status:    statusCode,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	})
}

func GenerateInternalServerErrorResponse(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, responses.APIResponse{
		Status:    http.StatusInternalServerError,
		Message:   message,
		Success:   false,
		Timestamp: time.Now(),
		Data:      gin.H{},
	})
}

func GenerateBadRequestResponse(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, responses.APIResponse{
		Status:    http.StatusBadRequest,
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
		Data:      gin.H{},
	})
}

func ExtractDuplicateKeyInfo(err error) string {
	errMsg := err.Error()
	startIdx := strings.Index(errMsg, "dup key: {")
	if startIdx == -1 {
		return "No duplicate key info found"
	}

	endIdx := strings.Index(errMsg[startIdx:], "}")
	if endIdx == -1 {
		return "Invalid duplicate key info"
	}
	length := 0
	for index := range errMsg {
		if index == startIdx {
			length++
		}
		if index == endIdx {
			break
		}
	}
	section := errMsg[startIdx+length : startIdx+endIdx]
	vals := strings.Split(section, ": {")
	res := strings.Split(vals[1], ":")

	key := strings.Trim(res[0], " ")
	return key
}
