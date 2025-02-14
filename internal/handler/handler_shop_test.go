package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testAlvtoShp/internal/service"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	internalErrors "testAlvtoShp/internal/errors"
	"testAlvtoShp/internal/models"
	"testAlvtoShp/internal/service/mocks"
)

func TestHandler_GetUserInfo(t *testing.T) {
	type mockBehavior func(m *mocks.MockShop, userId int)
	testTable := []struct {
		name                 string
		userId               int
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:   "Error getting user info",
			userId: 42,
			mockBehavior: func(m *mocks.MockShop, userId int) {
				m.EXPECT().
					GetUserInfo(gomock.Any(), userId).
					Return(models.InfoResponse{}, errors.New("db error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"errors": "Error getting user info"}`,
		},
		{
			name:   "Successful getting user info",
			userId: 42,
			mockBehavior: func(m *mocks.MockShop, userId int) {
				info := models.InfoResponse{
					Coins:     100,
					Inventory: []models.Item{},
					CoinHistory: models.CoinHistory{
						Received: []models.ReceivedTransaction{},
						Sent:     []models.SentTransaction{},
					},
				}
				m.EXPECT().
					GetUserInfo(gomock.Any(), userId).
					Return(info, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"coins":100,"inventory":[],"coinHistory":{"received":[],"sent":[]}}`,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockShop := mocks.NewMockShop(ctrl)
			if tc.mockBehavior != nil {
				tc.mockBehavior(mockShop, tc.userId)
			}

			h := NewHandler(&service.Service{
				Shop: mockShop,
			})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userId", tc.userId)
			req := httptest.NewRequest("GET", "/userinfo", nil)
			c.Request = req

			h.GetUserInfo(c)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			if tc.expectedResponseBody != "" {
				assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
			}
		})
	}
}

func TestHandler_BuyItem(t *testing.T) {
	type mockBehavior func(m *mocks.MockShop, userId int, item string)
	testTable := []struct {
		name                 string
		userId               int
		item                 string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:   "Incorrect item provided",
			userId: 42,
			item:   "sword",
			mockBehavior: func(m *mocks.MockShop, userId int, item string) {
				m.EXPECT().
					BuyItem(gomock.Any(), userId, item).
					Return(sql.ErrNoRows)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"errors": "Incorrect item"}`,
		},
		{
			name:   "No money for item",
			userId: 42,
			item:   "shield",
			mockBehavior: func(m *mocks.MockShop, userId int, item string) {
				m.EXPECT().
					BuyItem(gomock.Any(), userId, item).
					Return(internalErrors.NoMoney)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"errors": "No money for this item"}`,
		},
		{
			name:   "Generic error buying item",
			userId: 42,
			item:   "potion",
			mockBehavior: func(m *mocks.MockShop, userId int, item string) {
				m.EXPECT().
					BuyItem(gomock.Any(), userId, item).
					Return(errors.New("some error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"errors": "Error buying item"}`,
		},
		{
			name:   "Successful purchase",
			userId: 42,
			item:   "armor",
			mockBehavior: func(m *mocks.MockShop, userId int, item string) {
				m.EXPECT().
					BuyItem(gomock.Any(), userId, item).
					Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{}`,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockShop := mocks.NewMockShop(ctrl)
			if tc.mockBehavior != nil {
				tc.mockBehavior(mockShop, tc.userId, tc.item)
			}

			h := NewHandler(&service.Service{
				Shop: mockShop,
			})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userId", tc.userId)
			c.Params = gin.Params{{Key: "item", Value: tc.item}}
			req := httptest.NewRequest("POST", "/buy/"+tc.item, nil)
			c.Request = req

			h.BuyItem(c)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			if tc.expectedResponseBody != "" {
				assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
			}
		})
	}
}

func TestHandler_SendCoin(t *testing.T) {
	type mockBehavior func(m *mocks.MockShop, userId int, req models.SendCoinRequest)
	testTable := []struct {
		name                 string
		requestBody          string
		userId               int
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                 "Binding error",
			requestBody:          "invalid json",
			userId:               42,
			mockBehavior:         nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"errors": "error with binding request"}`,
		},
		{
			name:                 "Bad request data - missing fields",
			requestBody:          `{"amount":0,"toUser":"Alice"}`,
			userId:               42,
			mockBehavior:         nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"errors": "bad request data"}`,
		},
		{
			name:                 "Amount less than 1",
			requestBody:          `{"amount":-5,"toUser":"Alice"}`,
			userId:               42,
			mockBehavior:         nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"errors": "Amount must be greater than zero"}`,
		},
		{
			name:        "Incorrect user error",
			requestBody: `{"amount":10,"toUser":"Alice"}`,
			userId:      42,
			mockBehavior: func(m *mocks.MockShop, userId int, req models.SendCoinRequest) {
				m.EXPECT().
					SendCoin(gomock.Any(), userId, req).
					Return(sql.ErrNoRows)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"errors": "Incorrect user"}`,
		},
		{
			name:        "No money error",
			requestBody: `{"amount":10,"toUser":"Alice"}`,
			userId:      42,
			mockBehavior: func(m *mocks.MockShop, userId int, req models.SendCoinRequest) {
				m.EXPECT().
					SendCoin(gomock.Any(), userId, req).
					Return(internalErrors.NoMoney)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"errors": "No money for this operation"}`,
		},
		{
			name:        "Generic error sending coin",
			requestBody: `{"amount":10,"toUser":"Alice"}`,
			userId:      42,
			mockBehavior: func(m *mocks.MockShop, userId int, req models.SendCoinRequest) {
				m.EXPECT().
					SendCoin(gomock.Any(), userId, req).
					Return(errors.New("some error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"errors": "error with sending coin"}`,
		},
		{
			name:        "Success",
			requestBody: `{"amount":10,"toUser":"Alice"}`,
			userId:      42,
			mockBehavior: func(m *mocks.MockShop, userId int, req models.SendCoinRequest) {
				m.EXPECT().
					SendCoin(gomock.Any(), userId, req).
					Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{}`,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockShop := mocks.NewMockShop(ctrl)
			var reqObj models.SendCoinRequest
			err := json.Unmarshal([]byte(tc.requestBody), &reqObj)
			if err == nil && tc.mockBehavior != nil {
				tc.mockBehavior(mockShop, tc.userId, reqObj)
			}

			h := NewHandler(&service.Service{
				Shop: mockShop,
			})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userId", tc.userId)
			req := httptest.NewRequest("POST", "/sendcoin", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			h.SendCoin(c)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			if tc.expectedResponseBody != "" {
				assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
			}
		})
	}
}
