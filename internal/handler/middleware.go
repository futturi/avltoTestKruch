package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testAlvtoShp/internal/logger"
	"testAlvtoShp/internal/models"
)

func (h *Handler) CheckAuth(c *gin.Context) {
	log := logger.LoggerFromContext(c.Request.Context())
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		log.Errorw("Authorization header is empty")
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Missing token"})
		return
	}
	splittedHeader := strings.Split(tokenString, " ")
	if len(splittedHeader) != 2 {
		log.Errorw("Authorization header is invalid")
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Missing token"})
		return
	}

	userId, err := h.service.ExtractUserIDFromAccessToken(c.Request.Context(), splittedHeader[1])
	if err != nil {
		log.Errorw("Authorization header is invalid", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid token"})
		return
	}

	c.Set("userId", userId)
}
