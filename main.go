package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MoveRequest struct {
	X int `json:"x" binding:"required,min=0,max=2"`
	Y int `json:"y" binding:"required,min=0,max=2"`
}

type GameResponse struct {
	Board   [SIZE][SIZE]string `json:"board"`
	Status  string             `json:"status"` // "ongoing", "X wins", "O wins", "draw"
	Message string             `json:"message"`
}

func main() {
	r := gin.Default()

	r.POST("/move", func(c *gin.Context) {
		var req MoveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid move coordinates. x and y must be 0-2."})
			return
		}

		board, status, msg, err := Play(req.X, req.Y)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, GameResponse{
			Board:   board,
			Status:  status,
			Message: msg,
		})
	})

	r.POST("/reset", func(c *gin.Context) {
		board := ResetGame()
		c.JSON(http.StatusOK, gin.H{"message": "Game reset", "board": board})
	})

	r.Run(":8080")
}
