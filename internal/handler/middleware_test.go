package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"testAlvtoShp/internal/service"
	"testAlvtoShp/internal/service/mocks"
)

func TestHandler_CheckAuth(t *testing.T) {
	type mockBehavior func(mockAuth *mocks.MockAuth, token string)

	testTable := []struct {
		name                 string
		headerValue          string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
		expectedUserID       interface{}
	}{
		{
			name:                 "Missing Authorization header",
			headerValue:          "",
			mockBehavior:         nil,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"errors":"Missing token"}`,
			expectedUserID:       nil,
		},
		{
			name:                 "Invalid header format",
			headerValue:          "Bearer",
			mockBehavior:         nil,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"errors":"Missing token"}`,
			expectedUserID:       nil,
		},
		{
			name:        "Invalid token",
			headerValue: "Bearer invalidtoken",
			mockBehavior: func(mockAuth *mocks.MockAuth, token string) {
				mockAuth.EXPECT().
					ExtractUserIDFromAccessToken(gomock.Any(), token).
					Return(0, errors.New("invalid token"))
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"errors":"Invalid token"}`,
			expectedUserID:       nil,
		},
		{
			name:        "Successful authentication",
			headerValue: "Bearer validtoken",
			mockBehavior: func(mockAuth *mocks.MockAuth, token string) {
				mockAuth.EXPECT().
					ExtractUserIDFromAccessToken(gomock.Any(), token).
					Return(42, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: "",
			expectedUserID:       42,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mocks.NewMockAuth(ctrl)
			if tc.mockBehavior != nil {
				parts := strings.Split(tc.headerValue, " ")
				token := ""
				if len(parts) == 2 {
					token = parts[1]
				}
				tc.mockBehavior(mockAuth, token)
			}

			h := NewHandler(&service.Service{
				Auth: mockAuth,
			})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/check", nil)
			req.Header.Set("Authorization", tc.headerValue)
			c.Request = req

			h.CheckAuth(c)

			if tc.expectedStatusCode != http.StatusOK {
				assert.Equal(t, tc.expectedStatusCode, w.Code)
				assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
			} else {
				assert.Equal(t, tc.expectedStatusCode, w.Code)
				userID, exists := c.Get("userId")
				assert.True(t, exists, "userId должен быть установлен в контекст")
				assert.Equal(t, tc.expectedUserID, userID)
			}
		})
	}
}
