# Poke TUI (Go)

A Go-based Text User Interface (TUI) application using the PokeAPI to display Pokémon and their attributes.

## Requirements

- Go 1.20+

Install dependencies:
```bash
go mod tidy
```

## Usage
To run without building:
```bash
go run main.go
```

To build and run:
```bash
go build -o pokeapitui main.go
./pokeapitui
```

Within the TUI:
- Arrow keys to navigate or Vim, btw.
- Enter to select a Pokémon and view details.
- Press `q` to quit.
