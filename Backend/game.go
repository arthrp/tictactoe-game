package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/redis/go-redis/v9"
)

const SIZE = 3

type Player string
type Winner string

const (
	PlayerX    Player = "X"
	PlayerO    Player = "O"
	WinnerX    Winner = Winner(PlayerX)
	WinnerO    Winner = Winner(PlayerO)
	WinnerDraw Winner = "draw"
)

type GameState struct {
	Board    [SIZE][SIZE]string `json:"board"`
	GameOver bool               `json:"game_over"`
}

func GetGameState(apiKey string) (*GameState, error) {
	val, err := redisClient.Get(ctx, "game:"+apiKey).Result()
	if err == redis.Nil {
		// If key doesn't exist, return a new game state
		return &GameState{
			Board:    [SIZE][SIZE]string{},
			GameOver: false,
		}, nil
	} else if err != nil {
		return nil, err
	}

	var state GameState
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func SaveGameState(apiKey string, state *GameState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return redisClient.Set(ctx, "game:"+apiKey, data, 0).Err()
}

func checkWinner(board [SIZE][SIZE]string) Winner {
	// Check rows and columns
	for i := 0; i < SIZE; i++ {
		if board[i][0] != "" && board[i][0] == board[i][1] && board[i][1] == board[i][2] {
			return Winner(board[i][0])
		}
		if board[0][i] != "" && board[0][i] == board[1][i] && board[1][i] == board[2][i] {
			return Winner(board[0][i])
		}
	}

	// Check diagonals
	if board[0][0] != "" && board[0][0] == board[1][1] && board[1][1] == board[2][2] {
		return Winner(board[0][0])
	}
	if board[0][2] != "" && board[0][2] == board[1][1] && board[1][1] == board[2][0] {
		return Winner(board[0][2])
	}

	// Check draw
	isDraw := true
	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			if board[i][j] == "" {
				isDraw = false
				break
			}
		}
	}
	if isDraw {
		return WinnerDraw
	}

	return ""
}

func aiMove(board *[SIZE][SIZE]string) {
	// Simple AI: Pick a random empty spot
	type point struct{ x, y int }
	var emptySpots []point

	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			if board[i][j] == "" {
				emptySpots = append(emptySpots, point{i, j})
			}
		}
	}

	if len(emptySpots) > 0 {
		// Try to find a winning move first
		for _, p := range emptySpots {
			board[p.x][p.y] = "O"
			if checkWinner(*board) == "O" {
				return
			}
			board[p.x][p.y] = "" // backtrack
		}

		// Try to block opponent winning move
		for _, p := range emptySpots {
			board[p.x][p.y] = "X"
			if checkWinner(*board) == "X" {
				board[p.x][p.y] = "O" // Block
				return
			}
			board[p.x][p.y] = "" // backtrack
		}

		// Pick random
		rng := rand.Intn(len(emptySpots))
		move := emptySpots[rng]
		board[move.x][move.y] = "O"
	}
}

func ResetGame(apiKey string) ([SIZE][SIZE]string, error) {
	state := &GameState{
		Board:    [SIZE][SIZE]string{},
		GameOver: false,
	}
	if err := SaveGameState(apiKey, state); err != nil {
		return state.Board, err
	}
	return state.Board, nil
}

func Play(apiKey string, x, y int) ([SIZE][SIZE]string, string, string, error) {
	state, err := GetGameState(apiKey)
	if err != nil {
		return [SIZE][SIZE]string{}, "", "", err
	}

	if state.GameOver {
		state.Board = [SIZE][SIZE]string{}
		state.GameOver = false
	}

	if x < 0 || x >= SIZE || y < 0 || y >= SIZE {
		return state.Board, "", "", fmt.Errorf("coordinates out of bounds")
	}

	if state.Board[x][y] != "" {
		return state.Board, "", "", fmt.Errorf("cell already occupied")
	}

	// Player move
	state.Board[x][y] = "X"
	winner := checkWinner(state.Board)

	if winner != "" {
		state.GameOver = true
		msg := formatWinMessage(winner)
		if err := SaveGameState(apiKey, state); err != nil {
			return state.Board, msg, "Error saving game state", err
		}
		return state.Board, msg, "Game Over", nil
	}

	// AI move
	aiMove(&state.Board)
	winner = checkWinner(state.Board)

	if winner != "" {
		state.GameOver = true
		msg := formatWinMessage(winner)
		if err := SaveGameState(apiKey, state); err != nil {
			return state.Board, msg, "Error saving game state", err
		}
		return state.Board, msg, "Game Over", nil
	}

	if err := SaveGameState(apiKey, state); err != nil {
		return state.Board, "ongoing", "Error saving game state", err
	}

	return state.Board, "ongoing", "Your turn", nil
}

func formatWinMessage(winner Winner) string {
	if winner == WinnerDraw {
		return "It's a draw!"
	}
	return fmt.Sprintf("%s wins!", winner)
}
