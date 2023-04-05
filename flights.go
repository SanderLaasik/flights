package main

import (
	"flights/database"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/spf13/viper"
)

type ConnectionParams struct {
	From  string    `form:"from" binding:"required,len=3,alphanum"`
	To    string    `form:"to" binding:"required,nefield=From,len=3,alphanum"`
	Date  time.Time `form:"date" binding:"required" time_format:"2006/01/02"`
	Limit int       `form:"limit" binding:"number"`
}

var router *gin.Engine

func main() {
	setupConf()
	setupDatabase()
	setupRouter()

	router.Run(":8080")
}

func setupConf() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
}

func setupDatabase() {
	database.Setup()
}

func setupRouter() {
	router = gin.Default()
	router.GET("/airports", getAirports)
	router.GET("/connections", findConnections)
}

func getAirports(c *gin.Context) {
	airports, err := database.GetAirports()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, airports)
}

func findConnections(c *gin.Context) {
	var connectionParams ConnectionParams
	if err := c.ShouldBindWith(&connectionParams, binding.Query); err == nil {
		var limit int

		if c.Query("limit") != "" {
			limit, _ = strconv.Atoi(c.Query("limit"))
		} else {
			limit = 5
		}
		date, err := time.Parse("2006/01/02", c.Query("date"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		connections, err := database.FindConnections(date.Format("2006-01-02"), c.Query("from"), c.Query("to"), limit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, connections)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
