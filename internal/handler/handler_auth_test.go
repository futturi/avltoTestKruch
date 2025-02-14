package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"testAlvtoShp/internal/models"
	"testAlvtoShp/internal/service"
	"testAlvtoShp/internal/service/mocks"
)

func TestHandler_GetAuthToken(t *testing.T) {
	type mockBehavior func(s *mocks.MockAuth, req models.AuthRequest)

	testTable := []struct {
		name                 string
		inputBody            string
		inputRequest         models.AuthRequest
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "Invalid JSON",
			inputBody: `invalid json`,
			mockBehavior: func(s *mocks.MockAuth, req models.AuthRequest) {
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"errors":"Error in parsing body"}`,
		},
		{
			name:                 "Missing Username",
			inputBody:            `{"password": "test"}`,
			mockBehavior:         func(s *mocks.MockAuth, req models.AuthRequest) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"errors":"Incorrect body"}`,
		},
		{
			name:                 "Missing Password",
			inputBody:            `{"username": "test"}`,
			mockBehavior:         func(s *mocks.MockAuth, req models.AuthRequest) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"errors":"Incorrect body"}`,
		},
		{
			name:      "GetUserByUsername Error",
			inputBody: `{"username": "test", "password": "test"}`,
			inputRequest: models.AuthRequest{
				Username: "test",
				Password: "test",
			},
			mockBehavior: func(s *mocks.MockAuth, req models.AuthRequest) {
				s.EXPECT().GetUserByUsername(gomock.Any(), req.Username).
					Return(0, errors.New("db error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"errors":"Error in getting user by username"}`,
		},
		{
			name:      "CreateUser Error",
			inputBody: `{"username": "newuser", "password": "newpass"}`,
			inputRequest: models.AuthRequest{
				Username: "newuser",
				Password: "newpass",
			},
			mockBehavior: func(s *mocks.MockAuth, req models.AuthRequest) {
				s.EXPECT().GetUserByUsername(gomock.Any(), req.Username).
					Return(0, nil)
				s.EXPECT().CreateUser(gomock.Any(), req).
					Return(0, errors.New("create user failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"errors":"Error in creating user"}`,
		},
		{
			name:      "GenerateToken Error for New User",
			inputBody: `{"username": "newuser", "password": "newpass"}`,
			inputRequest: models.AuthRequest{
				Username: "newuser",
				Password: "newpass",
			},
			mockBehavior: func(s *mocks.MockAuth, req models.AuthRequest) {
				s.EXPECT().GetUserByUsername(gomock.Any(), req.Username).
					Return(0, nil)
				s.EXPECT().CreateUser(gomock.Any(), req).
					Return(42, nil)
				s.EXPECT().GenerateToken(42).
					Return("", errors.New("token generation failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"errors":"Error in generating token"}`,
		},
		{
			name:      "CheckPassword Error",
			inputBody: `{"username": "existinguser", "password": "wrongpass"}`,
			inputRequest: models.AuthRequest{
				Username: "existinguser",
				Password: "wrongpass",
			},
			mockBehavior: func(s *mocks.MockAuth, req models.AuthRequest) {
				s.EXPECT().GetUserByUsername(gomock.Any(), req.Username).
					Return(1, nil)
				s.EXPECT().CheckPassword(gomock.Any(), req).
					Return(false, errors.New("check password failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"errors":"Error in checking password"}`,
		},
		{
			name:      "Invalid Password",
			inputBody: `{"username": "existinguser", "password": "wrongpass"}`,
			inputRequest: models.AuthRequest{
				Username: "existinguser",
				Password: "wrongpass",
			},
			mockBehavior: func(s *mocks.MockAuth, req models.AuthRequest) {
				s.EXPECT().GetUserByUsername(gomock.Any(), req.Username).
					Return(1, nil)
				s.EXPECT().CheckPassword(gomock.Any(), req).
					Return(false, nil)
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"errors":"Invalid password"}`,
		},
		{
			name:      "GenerateToken Error for Existing User",
			inputBody: `{"username": "existinguser", "password": "correctpass"}`,
			inputRequest: models.AuthRequest{
				Username: "existinguser",
				Password: "correctpass",
			},
			mockBehavior: func(s *mocks.MockAuth, req models.AuthRequest) {
				s.EXPECT().GetUserByUsername(gomock.Any(), req.Username).
					Return(1, nil)
				s.EXPECT().CheckPassword(gomock.Any(), req).
					Return(true, nil)
				s.EXPECT().GenerateToken(1).
					Return("", errors.New("token generation failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"errors":"Error in generating token"}`,
		},
		{
			name:      "Success Existing User",
			inputBody: `{"username": "existinguser", "password": "correctpass"}`,
			inputRequest: models.AuthRequest{
				Username: "existinguser",
				Password: "correctpass",
			},
			mockBehavior: func(s *mocks.MockAuth, req models.AuthRequest) {
				s.EXPECT().GetUserByUsername(gomock.Any(), req.Username).
					Return(1, nil)
				s.EXPECT().CheckPassword(gomock.Any(), req).
					Return(true, nil)
				s.EXPECT().GenerateToken(1).
					Return("validtoken", nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"token":"validtoken"}`,
		},
		{
			name:      "Success New User",
			inputBody: `{"username": "newuser", "password": "newpass"}`,
			inputRequest: models.AuthRequest{
				Username: "newuser",
				Password: "newpass",
			},
			mockBehavior: func(s *mocks.MockAuth, req models.AuthRequest) {
				s.EXPECT().GetUserByUsername(gomock.Any(), req.Username).
					Return(0, nil)
				s.EXPECT().CreateUser(gomock.Any(), req).
					Return(42, nil)
				s.EXPECT().GenerateToken(42).
					Return("newusertoken", nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"token":"newusertoken"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockAuth(ctrl)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockService, testCase.inputRequest)
			}

			handler := NewHandler(&service.Service{
				Auth: mockService,
			})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("POST", "/auth", bytes.NewBufferString(testCase.inputBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			handler.GetAuthToken(c)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			if testCase.expectedResponseBody != "" {
				assert.JSONEq(t, testCase.expectedResponseBody, w.Body.String())
			}
		})
	}
}
