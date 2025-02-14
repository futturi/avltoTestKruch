package handler

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	internalErrors "testAlvtoShp/internal/errors"
	"testAlvtoShp/internal/logger"
	"testAlvtoShp/internal/models"
)

// @Summary GetUserInfo
// @Security ApiKeyAuth
// @Tags shop
// @Description get info 4 user
// @ID get-info-4-user
// @Produce json
// @Success 200 {object} models.InfoResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/info [get]
func (h *Handler) GetUserInfo(c *gin.Context) {
	log := logger.LoggerFromContext(c.Request.Context())
	userId := c.GetInt("userId")

	userInfo, err := h.service.GetUserInfo(c.Request.Context(), userId)

	if err != nil {
		log.Error("error with getting user info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Error getting user info",
		})

		return
	}
	log.Infow("got user info", "userId", userId)
	c.JSON(http.StatusOK, userInfo)
}

// @Summary BuyItem
// @Security ApiKeyAuth
// @Tags shop
// @Description buy item 4 user
// @ID buy-item-4-user
// @Produce json
// @Param item path models.ItemForBuy true "Item to purchase" models.ItemForBuy
// @Success 200 {object} nil
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/buy/{item} [get]
func (h *Handler) BuyItem(c *gin.Context) {
	log := logger.LoggerFromContext(c.Request.Context())
	userId := c.GetInt("userId")
	item := c.Param("item")

	err := h.service.BuyItem(c.Request.Context(), userId, item)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Errorw("incorrect item provided", zap.Error(err))
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Incorrect item",
			})
			return
		} else if errors.Is(err, internalErrors.NoMoney) {
			log.Errorw("incorrect item provided", zap.Error(err))
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "No money for this item",
			})
			return
		}
		log.Errorw("error buying item", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Error buying item",
		})
		return
	}

	log.Infow("buy item successfully", zap.String("item", item), zap.Int("userId", userId))
	c.JSON(http.StatusOK, gin.H{})
}

// @Summary SendCoin
// @Security ApiKeyAuth
// @Tags shop
// @Description send coin to user
// @ID send-coin-to-user
// @Produce json
// @Param input body models.SendCoinRequest true "account info"
// @Success 200 {object} nil
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/sendCoin [post]
func (h *Handler) SendCoin(c *gin.Context) {
	log := logger.LoggerFromContext(c.Request.Context())

	var req models.SendCoinRequest

	if err := c.ShouldBind(&req); err != nil {
		log.Errorw("error with binding request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "error with binding request", //todo const
		})
		return
	}

	if req.Amount == 0 || req.ToUser == "" {
		log.Errorw("bad request data provided")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad request data",
		})
		return
	}

	if req.Amount < 1 {
		log.Errorw("amount must be greater than zero", zap.Any("amount", req.Amount))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Amount must be greater than zero",
		})
		return
	}

	userId := c.GetInt("userId")

	err := h.service.SendCoin(c.Request.Context(), userId, req)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Errorw("incorrect user", zap.Error(err))
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Incorrect user",
			})
			return
		} else if errors.Is(err, internalErrors.NoMoney) {
			log.Errorw("noMoney for this operation", zap.Error(err))
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "No money for this operation",
			})
			return
		}
		log.Errorw("error with sending coin", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "error with sending coin",
		})
		return
	}

	log.Infow("send coin successfully", zap.Int("amount", req.Amount), zap.Int("userId", userId), zap.String("toUser", req.ToUser))

	c.JSON(http.StatusOK, gin.H{})
}
