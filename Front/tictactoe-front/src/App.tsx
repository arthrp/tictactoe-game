import { useState, useEffect } from 'react';
import './App.css';

type Board = string[][];

interface GameResponse {
  board: Board;
  status: string;
  message: string;
}

function App() {
  const [apiKey, setApiKey] = useState('');
  const [board, setBoard] = useState<Board>([
    ['', '', ''],
    ['', '', ''],
    ['', '', ''],
  ]);
  const [status, setStatus] = useState('ongoing');
  const [message, setMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchGameState = async (key: string) => {
    if (!key) return;
    setIsLoading(true);
    setError('');
    try {
      const res = await fetch('/api/state', {
        headers: {
          'x-api-key': key,
        },
      });
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || 'Failed to fetch game state');
      }
      const data: GameResponse = await res.json();
      setBoard(data.board);
      setStatus(data.status);
      if (data.message) setMessage(data.message);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleMove = async (x: number, y: number) => {
    if (!apiKey) {
      setError('Please enter an API Key');
      return;
    }
    if (board[x][y] !== '' || status !== 'ongoing') return;

    setIsLoading(true);
    setError('');
    try {
      const res = await fetch('/api/move', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-api-key': apiKey,
        },
        body: JSON.stringify({ x, y }),
      });
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || 'Failed to make move');
      }
      const data: GameResponse = await res.json();
      setBoard(data.board);
      setStatus(data.status);
      setMessage(data.message);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleReset = async () => {
    if (!apiKey) return;
    setIsLoading(true);
    setError('');
    try {
      const res = await fetch('/api/reset', {
        method: 'POST',
        headers: {
          'x-api-key': apiKey,
        },
      });
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || 'Failed to reset game');
      }
      const data = await res.json();
      setBoard(data.board);
      setStatus('ongoing');
      setMessage(data.message);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Fetch initial state when apiKey changes (debounced or on blur/enter could be better, 
  // but let's just do it when they stop typing or we can add a "Load" button. 
  // The requirement says "input field below the grid". 
  // I'll make the input control the key and maybe a button to "Connect" or just use the key for requests.
  // Ideally, we load state once a valid key is provided.
  // For simplicity, I'll add a button "Load Game" or just rely on the first move/action.
  // But to see the board state, we should probably fetch it.
  
  // Let's rely on a "Load / Refresh" button or just calling it on mount if we had a key.
  // Since key is user input, let's auto-fetch when key is valid format? No, let's keep it simple with a button "Connect" or just let them play.
  // Wait, if the user inputs the key, they might want to see the current state.
  // I'll add a "Load Game" button next to the input.

  return (
    <div className="App">
      <h1>Tic Tac Toe</h1>
      
      <div className="status">
        Status: {status}
        {message && <div style={{ marginTop: '5px', fontSize: '0.9em', color: '#666' }}>{message}</div>}
      </div>

      {error && <div style={{ color: 'red', marginBottom: '10px' }}>Error: {error}</div>}

      <div className="board">
        {board.map((row, rowIndex) =>
          row.map((cell, colIndex) => (
            <div
              key={`${rowIndex}-${colIndex}`}
              className="cell"
              onClick={() => handleMove(rowIndex, colIndex)}
            >
              {cell}
            </div>
          ))
        )}
      </div>

      <div className="input-container">
        <label htmlFor="apiKey">API Key:</label>
        <input
          id="apiKey"
          type="password"
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
          placeholder="Enter your UUID API Key"
        />
        <div style={{ display: 'flex', gap: '10px' }}>
            <button onClick={() => fetchGameState(apiKey)} disabled={isLoading || !apiKey}>
            Load Game
            </button>
            <button onClick={handleReset} disabled={isLoading || !apiKey}>
            Reset Game
            </button>
        </div>
      </div>
    </div>
  );
}

export default App;
