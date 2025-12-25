package main

import (
	"fmt"
	"math/rand"
	"sync"
)

const SIZE = 3

type Player string
type Winner string

const (
	PlayerX    Player = "X"
	PlayerO    Player = "O"
	WinnerX    Winner = Winner(PlayerX)
	WinnerY    Winner = Winner(PlayerO)
	WinnerDraw Winner = "draw"
)

var (
	board    [SIZE][SIZE]string
	mutex    sync.Mutex
	gameOver bool
)

func init() {
	resetBoard()
}

func resetBoard() {
	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			board[i][j] = ""
		}
	}
	gameOver = false
}

func checkWinner() Winner {
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

func aiMove() {
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
			if checkWinner() == "O" {
				return
			}
			board[p.x][p.y] = "" // backtrack
		}

		// Try to block opponent winning move
		for _, p := range emptySpots {
			board[p.x][p.y] = "X"
			if checkWinner() == "X" {
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

func ResetGame() [SIZE][SIZE]string {
	mutex.Lock()
	defer mutex.Unlock()
	resetBoard()
	return board
}

func Play(x, y int) ([SIZE][SIZE]string, string, string, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if gameOver {
		resetBoard()
	}

	if x < 0 || x >= SIZE || y < 0 || y >= SIZE {
		return board, "", "", fmt.Errorf("coordinates out of bounds")
	}

	if board[x][y] != "" {
		return board, "", "", fmt.Errorf("cell already occupied")
	}

	// Player move
	board[x][y] = "X"
	winner := checkWinner()

	if winner != "" {
		gameOver = true
		msg := formatWinMessage(winner)
		return board, msg, "Game Over. Next move will start a new game.", nil
	}

	// AI move
	aiMove()
	winner = checkWinner()

	if winner != "" {
		gameOver = true
		msg := formatWinMessage(winner)
		return board, msg, "Game Over. Next move will start a new game.", nil
	}

	return board, "ongoing", "Your turn", nil
}

func formatWinMessage(winner Winner) string {
	if winner == WinnerDraw {
		return "It's a draw!"
	}
	return fmt.Sprintf("%s wins!", winner)
}
