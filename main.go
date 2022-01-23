package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	//  os.Setenv("PORT", "YOUR PORT")
	//	os.Setenv("API_KEY", "YOUR API KEY")
}

func main() {
	if os.Getenv("PORT") == "" {
		log.Fatal("port cannot be empty!")
	}
	if os.Getenv("API_KEY") == "" {
		log.Fatal("api key cannot be empty!")
	}
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.GET("/balance", HandleGetBalance)
	log.Fatal(r.Run(":" + os.Getenv("PORT")))
}

func HandleGetBalance(c *gin.Context) {
	network := c.Query("network")
	address := c.Query("address")
	if network == "" || address == "" {
		c.JSON(http.StatusBadRequest, "network and address paramethers are required")
		return
	}
	query := make(map[string]string)
	query["query"] = fmt.Sprintf(`query {ethereum(network: %s) { address(address: {is: "%s"}) {balances {  currency {	address	symbol	tokenType } value}}} }`, network, address)
	jsonQuery, err := json.Marshal(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	request, err := http.NewRequest("POST", "https://graphql.bitquery.io/", bytes.NewBuffer(jsonQuery))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	request.Header.Set("X-API-KEY", os.Getenv("API_KEY"))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if response.StatusCode == 200 {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		var respData interface{}
		err = json.Unmarshal(data, &respData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, respData)
		return
	}
	c.JSON(response.StatusCode, response.Body)
}
