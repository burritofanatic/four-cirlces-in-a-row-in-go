package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

type Room struct {
	Name        string 					`bson:"name" json:"name"`
	Clients     map[string]*Client
	Count       int                  	`bson:"count" json:"count"`
	Index       int                     `bson:"index" json:"index"`
	IsNew       bool                    `bson:"is_new" json:"is_new"`
	Turn        int                     `bson:"turn" json:"turn"`
	Board       [][]int              	`bson:"board" json:"board"`
	FirstScore  int                    	`bson:"first_score" json:"first_score"`
	SecondScore int                  	`bson:"second_score" json:"second_score"`
	InPlay      bool                	`bson:"in_play" json:"in_play"`
	Winner      int                    	`bson:"winner" json:"winner"`
	PlayerId    map[string]string    	`bson:"player_id" json:"player_id"`
	CurrentGame *Game                  	
	Games       []string              	`bson:"games" json:"games"`
	ID          bson.ObjectId			`bson:"_id" json:"id"`
	Tied		bool					`bson:"tied" json:"tied"`
	RematchRequest map[string]bool
	BlackList 	[]string				`bson:"blacklist" json:"blacklist"`
}

type Movement struct {
    Room 		string
    Turn 		int
    Board 		[][]int
    ScoreOne 	int
    ScoreTwo 	int
    InPlay 		bool
    Winner		int
    Message		string
	PlayerId    map[string]string
    TurnCount	int
    WinnerId 	string
    Tied		bool
    RematchRequest map[string]bool
}


func (r *Room) Join(conn *websocket.Conn, deviceId string, game Game) int {
	r.Index++

	r.Clients[deviceId] = NewClient(conn)
	r.RematchRequest = make(map[string]bool)

	client := r.Clients[deviceId]

	// Check to see if we're going to resurrect the game into the room.
	if game.gameExists() {
		r.CurrentGame = &game // Set the game the room and assign the player positions.
		fmt.Println("Existing game playerIds: ", r.PlayerId)


	} else {
		// If there is no preexisting game, handle the room anew, and update the database of the values.

		// When someone joins a room that already has game play, make sure that the winner is reset.
		r.Winner = 0

		// If both seats are empty, put the client in position one.
		if len(r.PlayerId) == 0 {
			client.player = 1
			r.PlayerId["1"] = deviceId

		} else {


			// When we reach here, this client will be assigned to the second seat.
			emptySeatString := FindEmptySeat(r.PlayerId)

			fmt.Println("Finding the empty seat... Empty seat is: ", emptySeatString)

			client.player, _ = strconv.Atoi(emptySeatString)
			r.PlayerId[emptySeatString] = deviceId

			fmt.Println("PlayerId is now: ", r.PlayerId)


			r.IsNew = false
			r.InPlay = true

			if r.CurrentGame != nil {
				if r.CurrentGame.TurnCount > 0 {
					// Only reset the game, and create a new one if the turn of the previous game is greater than zero.
					// This is for the instance in which one player remains in the room, and someone new joins, so that they can
					// start anew.

					r.ResetGame()
				} else {
					// Here, we know that the current game is new is turn out is zero
					// Set the second player in the current game
					r.CurrentGame.PlayerTwo = deviceId
				}
			}
		}

		r.Count++

		if r.CurrentGame == nil {
			// If the existing room that the user is joining does not have a game, create one.
			r.CurrentGame = r.CreateNewGameWithStartingPositions()
		} else {
			// Call to update the database.
			GameDaoInstance().UpdateGame(r.CurrentGame)
		}

		// Call to update the database.
		RoomDaoInstance().UpdateRoom(r)
	}

	// This responds to the connected client only of their own joining.
	r.HandleJoin(deviceId)

	// This is the generic room broadcast of moves, and in this case, of the join.
	r.HandleMove("A user has joined the table.")

	return r.Index
}

func FindEmptySeat(playerIdMap map[string]string) string {
	if _, ok := playerIdMap["1"]; ok {
		return "2"
	}
	return "1"
}

func (r *Room) CreateNewGameWithStartingPositions() *Game {
	var board [][][]int
	board = append(board, r.Board)
	game := NewGame(r)

	// Save the game to db
	err := GameDaoInstance().CreateGame(game)
	if err != nil {
		fmt.Println(err.Error())
	}
	return game
}

func (r *Room) CheckBoard(player int) bool {

	boardWidth := len(r.Board)
	boardHeight := len(r.Board[0])

	// Horizontal
	for j := 0; j < boardHeight - 3; j++  {
		for i := 0; i < boardWidth; i++ {
			if r.Board[i][j] == player && r.Board[i][j + 1] == player && r.Board[i][j + 2] == player && r.Board[i][j + 3] == player {
				return true
			}
		}
	}

	// Vertical
	for i := 0; i < boardWidth - 3; i++  {
		for j := 0; j < boardHeight; j++ {
			if r.Board[i][j] == player && r.Board[i + 1][j] == player && r.Board[i + 2][j] == player && r.Board[i + 3][j] == player {
				return true
			}
		}
	}

	// Ascending Diagonal
	for i := 3; i < boardWidth; i++  {
		for j := 0; j < boardHeight - 3; j++ {
			if r.Board[i][j] == player && r.Board[i - 1][j + 1] == player && r.Board[i - 2][j + 2] == player && r.Board[i - 3][j + 3] == player {
				return true
			}
		}
	}

	// Descending Diagonal
	for i := 3; i < boardWidth; i++  {
		for j := 3; j < boardHeight; j++ {
			if r.Board[i][j] == player && r.Board[i - 1][j - 1] == player && r.Board[i - 2][j - 2] == player && r.Board[i - 3][j - 3] == player {
				return true
			}
		}
	}

	return false
}


// If game is in play, and the user leaves the room, we force a loss
func (r *Room)ForceLoss(resigner int) {

	if r.InPlay {
		// Determine the client to leave
		leaving := resigner
		if len(r.Clients) == 2 {

			// Declare new Winner
			winner := OtherPlayer(leaving)
			r.DeclareWinner(winner)

			// Update the Tables.
			GameDaoInstance().UpdateGame(r.CurrentGame)
			RoomDaoInstance().UpdateRoom(r)
			r.HandleMove("An opponent has resigned and left.")
		}
	}
}

func (r *Room) DeclareWinner(winner int) {
	// We have a Winner
	r.Winner = winner
	r.CurrentGame.Winner = strconv.Itoa(winner)
	r.CurrentGame.WinnerId = r.PlayerId[strconv.Itoa(winner)]

	// Set the for the player
	if winner == 1 {
		r.FirstScore++
	} else {
		r.SecondScore++
	}

	// Set the game to a not ready state
	r.InPlay = false
}

func (r *Room) DeclareTie() {
	// We have a Winner
	r.Winner = 3
	r.CurrentGame.Winner = "3"
	r.CurrentGame.WinnerId = ""

	r.Tied = true
	r.InPlay = false
}

func OtherPlayer(player int) int {
	if player == 1 {
		return 2
	}
	return 1
}


/* Removes client from room */
func (r *Room) Leave(deviceId string) {
	r.Count--

	delete(r.Clients, deviceId)
}

/* Removes client from room */
func (r *Room) Blacklist(deviceId string) {
	r.BlackList = append(r.BlackList, deviceId)
	RoomDaoInstance().UpdateRoom(r)
}

/* Broadcast to every client */
func (r *Room) BroadcastAll(msg []byte) {
	for _, client := range r.Clients {
		client.WriteMessage(msg)
	}
}

/* Handle messages */
func (r *Room) HandleMsg(d string) {

	for {
		if r.Clients[d] == nil {
			break
		}

		out := <-r.Clients[d].out
		r.BroadcastAll(out.msg)
	}
}

/* Handle new move */
func (r *Room) HandleMove(m string) {
	for _, client := range r.Clients {

		data := Movement{ r.Name, r.Turn, r.Board,
						r.FirstScore, r.SecondScore, r.InPlay, r.Winner, m,
				r.PlayerId, r.CurrentGame.TurnCount, r.CurrentGame.WinnerId,
					r.Tied, r.RematchRequest}
		client.WriteJSON(data)
	}
}

func (r *Room) HandleJoin(deviceId string) {
	data := Movement{ r.Name, r.Turn, r.Board,
		r.FirstScore, r.SecondScore, r.InPlay, r.Winner, "",
		r.PlayerId, r.CurrentGame.TurnCount,
		r.CurrentGame.WinnerId, r.Tied, r.RematchRequest}
	r.Clients[deviceId].WriteJSON(data)
}

func (r *Room) ResetGame() {
	// Reset game

	// Clear out rematch request
	r.RematchRequest = make(map[string]bool)

	// Zero out the Board
	for i := 0; i < len(r.Board); i++ {
		for j := 0; j < len(r.Board[0]); j++ {
			r.Board[i][j] = 0
		}
	}

	r.Turn = 1
	r.InPlay = true
	r.Tied = false

	// Reset the Winner
	r.Winner = 0

	// Reset the rematch request
	r.RematchRequest = make(map[string]bool)

	// New Game
	var board [][][]int
	board = append(board, r.Board)
	game := NewGame(r)
	r.CurrentGame = game

	// Add the game to the database.
	GameDaoInstance().CreateGame(game)
}

/* Constructor */
func NewRoom(name string) *Room {
	room := new(Room)
	room.Name = name
	room.Clients = make(map[string]*Client)
	room.Count = 0
	room.Index = 0
	room.IsNew = true
	room.Turn = 1
	room.Board = [][]int{
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 0, 0, 0, 0} ,
	}
	room.FirstScore = 0
	room.SecondScore = 0
	room.InPlay = false
	room.Winner = 0
	room.PlayerId = make(map[string]string)
	room.ID = bson.NewObjectId()
	room.Tied = false
	room.RematchRequest = make(map[string]bool)
	room.BlackList = make([]string, 0)
	return room
}