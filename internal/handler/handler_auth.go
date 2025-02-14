package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"testAlvtoShp/internal/logger"
	"testAlvtoShp/internal/models"
)

// @Summary GetAuthToken
// @Tags auth
// @Description create/login account 4 user
// @ID create/login-account-user
// @Accept json
// @Produce json
// @Param input body models.AuthRequest true "account info"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth [post]
func (h *Handler) GetAuthToken(c *gin.Context) {
	log := logger.LoggerFromContext(c.Request.Context())

	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Error in parsing body",
		})
		return
	}

	if req.Password == "" || req.Username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Incorrect body",
		})
		return
	}

	userId, err := h.service.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		log.Errorw("GetUserByUsername", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Error in getting user by username",
		})
		return
	}

	if userId == 0 {
		var id int
		id, err = h.service.CreateUser(c.Request.Context(), req)
		if err != nil {
			log.Errorw("CreateUser", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Error in creating user",
			})
			return
		}

		userId = id
	} else {
		var isValid bool
		isValid, err = h.service.CheckPassword(c.Request.Context(), req)

		if err != nil {
			log.Errorw("CheckPassword", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Error in checking password",
			})
			return
		}
		if !isValid {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid password",
			})
			return
		}
	}

	token, err := h.service.GenerateToken(userId)
	if err != nil {
		log.Errorw("GenerateToken", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Error in generating token",
		})
		return
	}

	log.Infow("user authorize", zap.Int("user_id", userId))

	c.JSON(http.StatusOK, models.AuthResponse{Token: token})
}
