package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type MoveRequest struct {
	X *int `json:"x" binding:"required,min=0,max=2"`
	Y *int `json:"y" binding:"required,min=0,max=2"`
}

type GameResponse struct {
	Board   [SIZE][SIZE]string `json:"board"`
	Status  string             `json:"status"` // "ongoing", "X wins", "O wins", "draw"
	Message string             `json:"message"`
}

func main() {
	// Initialize Redis and load API keys
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	if err := InitRedis(redisAddr); err != nil {
		log.Fatalf("Could not initialize Redis: %v", err)
	}

	r := gin.Default()

	// Apply AuthMiddleware to all routes
	r.Use(AuthMiddleware())

	r.GET("/state", func(c *gin.Context) {
		apiKey := c.GetString("apiKey")
		state, err := GetGameState(apiKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get game state"})
			return
		}

		status := "ongoing"
		if state.GameOver {
			winner := checkWinner(state.Board)
			if winner != "" {
				status = formatWinMessage(winner)
			}
		}

		c.JSON(http.StatusOK, GameResponse{
			Board:  state.Board,
			Status: status,
		})
	})

	r.POST("/move", func(c *gin.Context) {
		apiKey := c.GetString("apiKey")
		var req MoveRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid move coordinates. x and y must be 0-2."})
			return
		}

		board, status, msg, err := Play(apiKey, *req.X, *req.Y)
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
		apiKey := c.GetString("apiKey")
		board, err := ResetGame(apiKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset game"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Game reset", "board": board})
	})

	r.Run(":8080")
}
