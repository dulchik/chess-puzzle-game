## Chess Engine (Stockfish)

This project uses Stockfish for AI play.

### Setup
1. Download Stockfish from:
    https://stockfishchess.org/download/

2. Extract the binary into: 
    engine/stockfish

3. Ensure it is executable:
    ```bash
    chmod +x engine/stockfish/stockfish*
    ```

4. Update the path in main.go if needed

chess/
├── main.go // bootstrapping only
├── game.go // Game struct + Update loop
├── board.go // board rendering
├── input.go // mouse / keyboard logic
├── moves.go // move list + history
├── clock.go // chess clock
├── ai.go // stockfish wrapper
├── menu.go // start menu UI
└── promotion.go // promotion picker
