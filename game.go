package main

import (

	"gopkg.in/mgo.v2/bson"


)

type Game struct {
	ID				bson.ObjectId	`bson:"_id" json:"id"`
	PlayerOne		string			`bson:"player_one" json:"player_one"` // These will be the deviceIds
	PlayerTwo		string			`bson:"player_two" json:"player_two"`
	BoardStates		[][][]int		`bson:"board_states" json:"board_states"`
	Winner			string 			`bson:"winner" json:"winner"` // player 1 or 2
	WinnerId		string 			`bson:"winner_id" json:"winner_id"` // deviceId
	Room			string			`bson:"room" json:"room"`
	TurnCount		int				`bson:"turn_count" json:"turn_count"`
}

func (g *Game) gameExists() bool {
	return g.ID != ""
}

func NewGame(room *Room) *Game {
	game := new(Game)
	game.ID = bson.NewObjectId()
	game.PlayerOne = room.PlayerId["1"]
	game.PlayerTwo = room.PlayerId["2"]

	var board [][][]int
	board = append(board, room.Board)
	game.BoardStates = [][][]int{{
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 0, 0, 0, 0} ,
	}}
	game.Winner = ""
	game.WinnerId = ""
	game.Room = room.Name

	// Whenever a new game is created, add it to the Room Table
	currentGames := room.Games
	room.Games = append(currentGames, game.ID.Hex())

	RoomDaoInstance().UpdateRoom(room)
	game.TurnCount = 0

	return game
}

