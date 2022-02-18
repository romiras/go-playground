package game

import "math/rand"

const (
	BattleshipMapSize uint = 10
	MaxPlayers        int  = 2
)

type (
	Coord struct {
		x uint
		y uint
	}

	PlayerID    string
	ShipID      string
	ShipDict    map[ShipID]Ship
	PlayersDict map[PlayerID]Player

	BattleshipCell struct {
		shipID *ShipID
	}

	Player struct {
		ID        PlayerID
		myArmy    BattleshipMap
		enemyArmy BattleshipMap
	}

	Ship struct {
		ID           string
		isSink       bool
		size         uint
		cellsDamaged uint
	}

	BattleshipMap struct {
		playerID PlayerID
		cells    [BattleshipMapSize][BattleshipMapSize]BattleshipCell
		shipDict ShipDict
		// lastFire *Coord
	}

	Game struct {
		playersDict PlayersDict
		currentTurn PlayerID
		winner      *PlayerID
		isOver      bool
	}
)

func NewGame(players []PlayerID) *Game {
	idx := rand.Intn(len(players))

	playersDict := make(PlayersDict)
	for _, playerID := range players {
		player := NewPlayer(playerID)
		playersDict[playerID] = *player
	}

	return &Game{
		playersDict: playersDict,
		isOver:      false,
		currentTurn: players[idx],
	}
}

func (game *Game) getCurrentTurn() PlayerID {
	return ""
}

func (game *Game) Play() {
	playerID := game.currentTurn
	for !game.isOver {
		player := game.playersDict[playerID]

		coord := Coord{
			x: 0,
			y: 0,
		}

		success := player.Fire(coord)
		if success {

		} else {
			game.currentTurn = game.getCurrentTurn()
		}

		game.isOver = true
	}
	game.winner = &playerID
}

func NewPlayer(id PlayerID) *Player {
	myArmy := BattleshipMap{
		playerID: id,
		cells:    [BattleshipMapSize][BattleshipMapSize]BattleshipCell{},
		shipDict: make(ShipDict),
	}

	enemyArmy := BattleshipMap{
		playerID: id,
		cells:    [BattleshipMapSize][BattleshipMapSize]BattleshipCell{},
		shipDict: make(ShipDict),
	}

	return &Player{
		ID:        id,
		myArmy:    myArmy,
		enemyArmy: enemyArmy,
	}
}

func (player *Player) Fire(coord Coord) bool {
	damaged := false

	return damaged
}
