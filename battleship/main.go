package main

import "github.com/romiras/battleship/game"

func main() {
	game := game.NewGame([]game.PlayerID{"A", "B"})
	game.Play()
}
