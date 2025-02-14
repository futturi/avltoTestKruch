package test

import (
	"context"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testAlvtoShp/internal/config"
	"testAlvtoShp/internal/handler"
	"testAlvtoShp/internal/models"
	"testAlvtoShp/internal/service"
	"testAlvtoShp/internal/store"
	"testing"
)

type IntegrationTestSuite struct {
	suite.Suite
	server *httptest.Server
	client *http.Client
	db     *sqlx.DB

	tokenA string
	tokenB string
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	var err error

	suite.db, err = sqlx.Connect("postgres", "postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable")
	suite.Require().NoError(err)

	cfg := &config.Config{
		JwtSecret: "testsecret",
	}

	s := service.NewService(&store.Store{
		Auth: store.NewAuthStore(suite.db),
		Shop: store.NewShopStore(suite.db),
	}, cfg)

	h := handler.NewHandler(s)

	router := h.InitRoutes(context.Background())
	suite.server = httptest.NewServer(router)
	suite.client = suite.server.Client()
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.server.Close()
	err := suite.db.Close()
	suite.Require().NoError(err)
}

func (suite *IntegrationTestSuite) TestAuthAndGetInfo() {
	reqBody := `{"username": "userA", "password": "passA"}`
	resp, err := suite.client.Post(suite.server.URL+"/api/auth", "application/json", strings.NewReader(reqBody))
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var authResp models.AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	suite.Require().NoError(err)

	err = resp.Body.Close()
	suite.Require().NoError(err)

	suite.NotEmpty(authResp.Token)
	suite.tokenA = authResp.Token

	req, err := http.NewRequest("GET", suite.server.URL+"/api/info", nil)
	suite.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer "+suite.tokenA)

	resp, err = suite.client.Do(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var infoResp models.InfoResponse
	err = json.NewDecoder(resp.Body).Decode(&infoResp)
	suite.Require().NoError(err)
	err = resp.Body.Close()

	suite.Require().NoError(err)
	suite.Equal(0, infoResp.Coins)
	suite.Empty(infoResp.Inventory)
	suite.Empty(infoResp.CoinHistory.Received)
	suite.Empty(infoResp.CoinHistory.Sent)
}
