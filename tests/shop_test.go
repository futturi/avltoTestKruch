package test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testAlvtoShp/internal/models"
)

func (suite *IntegrationTestSuite) TestBuyItem() {
	_, err := suite.db.Exec("UPDATE users SET coins = 100 WHERE username = $1", "userA")
	suite.Require().NoError(err)

	req, err := http.NewRequest("GET", suite.server.URL+"/api/buy/t-shirt", nil)
	suite.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer "+suite.tokenA)

	resp, err := suite.client.Do(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusOK, resp.StatusCode)
	err = resp.Body.Close()
	suite.Require().NoError(err)

	req, err = http.NewRequest("GET", suite.server.URL+"/api/info", nil)
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
	suite.Equal(20, infoResp.Coins)

	found := false
	for _, item := range infoResp.Inventory {
		if item.Type == "t-shirt" {
			found = true
			suite.GreaterOrEqual(item.Quantity, 1)
		}
	}
	suite.True(found, "t-shirt should be present in inventory")
}

func (suite *IntegrationTestSuite) TestSendCoin() {
	reqBody := `{"username": "userB", "password": "passB"}`

	resp, err := suite.client.Post(suite.server.URL+"/api/auth", "application/json", strings.NewReader(reqBody))
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var authResp models.AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	suite.Require().NoError(err)
	err = resp.Body.Close()

	suite.Require().NoError(err)
	suite.NotEmpty(authResp.Token)
	suite.tokenB = authResp.Token

	_, err = suite.db.Exec("UPDATE users SET coins = 100 WHERE username = $1", "userA")
	suite.Require().NoError(err)

	sendReqBody := `{"toUser": "userB", "amount": 50}`
	req, err := http.NewRequest("POST", suite.server.URL+"/api/sendCoin", strings.NewReader(sendReqBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.tokenA)
	resp, err = suite.client.Do(req)

	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	err = resp.Body.Close()
	suite.Require().NoError(err)

	req, err = http.NewRequest("GET", suite.server.URL+"/api/info", nil)
	suite.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer "+suite.tokenA)

	resp, err = suite.client.Do(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var infoA models.InfoResponse
	err = json.NewDecoder(resp.Body).Decode(&infoA)
	suite.Require().NoError(err)
	err = resp.Body.Close()
	suite.Require().NoError(err)
	suite.Equal(50, infoA.Coins)

	req, err = http.NewRequest("GET", suite.server.URL+"/api/info", nil)
	suite.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer "+suite.tokenB)

	resp, err = suite.client.Do(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var infoB models.InfoResponse
	err = json.NewDecoder(resp.Body).Decode(&infoB)
	suite.Require().NoError(err)

	err = resp.Body.Close()
	suite.Require().NoError(err)
	suite.Equal(50, infoB.Coins)
}
