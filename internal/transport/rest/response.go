package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorResponseBody struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

type successResponseBody struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type messageResponseType struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type paginatedResponse struct {
	Data       interface{} `json:"data"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

func successResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, successResponseBody{
		Status: "success",
		Data:   data,
	})
}

func errorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, errorResponseBody{
		Status:  "error",
		Message: message,
		Code:    statusCode,
	})
}

func messageResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, messageResponseType{
		Status:  "success",
		Message: message,
	})
}

func paginatedSuccessResponse(c *gin.Context, data interface{}, totalCount, page, pageSize int) {
	totalPages := totalCount / pageSize
	if totalCount%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, paginatedResponse{
		Data:       data,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func createdResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, successResponseBody{
		Status: "success",
		Data:   data,
	})
}

func noContentResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func badRequestResponse(c *gin.Context, message string) {
	errorResponse(c, http.StatusBadRequest, message)
}

func unauthorizedResponse(c *gin.Context) {
	errorResponse(c, http.StatusUnauthorized, "требуется авторизация")
}

func forbiddenResponse(c *gin.Context, message ...string) {
	msg := "доступ запрещен"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	errorResponse(c, http.StatusForbidden, msg)
}

func notFoundResponse(c *gin.Context, message string) {
	errorResponse(c, http.StatusNotFound, message)
}

func internalServerErrorResponse(c *gin.Context) {
	errorResponse(c, http.StatusInternalServerError, "внутренняя ошибка сервера")
}
