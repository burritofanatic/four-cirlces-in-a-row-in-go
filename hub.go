package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Controls rooms, and as an all around coordinator.
type Hub struct {
	hub      map[string]*Room
	upgrader websocket.Upgrader
	db 		DB
}

type moveStruct struct {
	Col int `json:"col"`
	Player int `json:"player"`
	Room string `json:"room"`
	PlayerId string `json:"PlayerId"`
}

type ChatMessage struct {
	Message string `json:"message"`
	Room string `json:"room"`
	PlayerId string `json:"PlayerId"`
}

type Rematch struct {
	Player int `json:"player"`
	Room string `json:"room"`
	PlayerId string `json:"PlayerId"`
}

type Resign struct {
	Player int `json:"player"`
	Room string `json:"room"`
	PlayerId string `json:"PlayerId"`
}

type Leave struct {
	Player int `json:"player"`
	Room string `json:"room"`
	PlayerId string `json:"PlayerId"`
}

type GameHistory struct {
	PlayerId string `json:"PlayerId"`
}

func (h *Hub) CreateRoom(name string) *Room {
	// Hash the "Name"
	hash := sha1.New()
	hash.Write([]byte(name))
	bs := hex.EncodeToString(hash.Sum(nil))

	if _, ok := h.hub[bs]; !ok {
		room := NewRoom(bs)
		h.hub[bs] = room

		err := RoomDaoInstance().CreateRoom(room)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return h.hub[bs]
}

// Iterate through all rooms, and find one with only one person. If none exists,
// then create a new room.

func (h *Hub) FindSeat(d string) (*Room, Game) {

	/* 	Handle resurrecting a game. A when a websocket connection disconnects, which is common,
	   	rely on an incomplete in play game, and return that old seat to the user without making
		a new room from scratch.
	 */

	game, err := GameDaoInstance().FindIncompleteGameForDevice(d)

	if err == nil && game.ID != "" {
		room, err := RoomDaoInstance().FindRoom(game.Room)

		if err == nil && room != nil {
			if h.hub[room.Name] != nil {
				return h.hub[room.Name], game
			} else {
				h.hub[room.Name] = room
				// Because this room is no longer in memory, clear out the clients.
				room.Clients = make(map[string]*Client)
				return room, game
			}
		}
	}

	log.Println("Finding a seat.")

	for _, room := range h.hub {
		/* Iterate through to check if the room has a seat available, that the game is not currently in play, and the
		 * existing deviceId is not already in.
		 */
		if len(room.PlayerId) < 2 && room.InPlay == false && !containsPlayerId(d, room) && !isBlackListed(d, room) {

			return room, game
		}
	}

	// If we make it here, create the room anew.
	log.Println("No seats found; creating a room anew.")
	return h.CreateRoom(time.Now().String()), game
}

func isBlackListed(d string, r *Room) bool {
	for _, b := range r.BlackList {
		if b == d {
			return true
		}
	}
	return false
}

func containsPlayerId(d string, r *Room) bool {
	for _, device := range r.PlayerId {
		if device == d {
			return true
		}
	}
	return false
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	deviceId := params["deviceId"]

	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	defer c.Close()

	log.Println("Join request received from: ", deviceId)

	room, game := h.FindSeat(deviceId) // This returns the first "random (order not guaranteed seat)"
	room.Join(c, deviceId, game)

	/* Reads from the client's out bound channel and broadcasts it */
	go room.HandleMsg(deviceId)

	go room.Clients[deviceId].WritePump()

	/* Reads from client and if this loop breaks then client disconnected. */
	room.Clients[deviceId].ReadLoop()

	room.Leave(deviceId)

	if len(room.Clients) == 0 {
		// If the room is empty, remove the room from memory entirely.
		if _, ok := h.hub[room.Name]; ok {
			delete(h.hub, room.Name)
			fmt.Println("Deleting Room from memory: ", room.Name)
		}
	}
}

func (h *Hub) HandleMessage(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var chat ChatMessage
	err := decoder.Decode(&chat)

	if err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}

	message := chat.Message
	roomhash := chat.Room
	room := h.hub[roomhash]
	playerId := chat.PlayerId

	var playerPosition string
	for k, v := range room.PlayerId {
		if v == playerId {
			playerPosition = k
		}
	}

	go room.HandleMove("Player " + playerPosition + ": " + message)

	defer r.Body.Close()
}

func (h *Hub) HandleMove(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var move moveStruct
	err := decoder.Decode(&move)

	if err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}

	col := move.Col
	player := move.Player
	roomhash := move.Room
	room := h.hub[roomhash]
	playerId := move.PlayerId

	if room == nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "The room does not exist. It's possible your connection was disconnected. Please background the app and start again.", 400)
		return
	}

	if room.InPlay == false {
		log.Printf("ERROR: %s", err)
		http.Error(w, "The game is over, you cannot submit a move.", 400)
		return
	}

	if col > len(room.Board[0]) {
		log.Printf("ERROR: %s", err)
		http.Error(w, "The move was out of bounds.", 400)
		return
	} else {

		fmt.Println("Players: ", room.PlayerId)

		originalBoardStates := room.CurrentGame.BoardStates

		duplicate := make([][][]int, len(originalBoardStates))
		for i := range originalBoardStates {
			duplicate[i] = make([][]int, len(originalBoardStates[i]))

			for j := range originalBoardStates[i] {
				duplicate[i][j] = make([]int, len(originalBoardStates[i][j]))
				copy(duplicate[i][j], originalBoardStates[i][j])
			}
		}

		// Validate that the position is question is empty, and that player is in Turn.
		if isTurn(player, room, playerId) {
			// Amend the 2D array

			// Note, in the game of connect four, the item is dropped.

			walker := 0

			// Check if in penultimate spot
			//TODO find a way to simplify keep board board state up to date and check by column [n] and just length
			if room.Board[walker + 1][col] == 0 { // Check to see if we can fall through
				for room.Board[walker][col] == 0 {
					if walker < len(room.Board) - 1 {
						walker++

						if walker == len(room.Board) - 1 {
							// We have reached the end of the Board.
							break
						} else if room.Board[walker + 1][col] != 0 {
							break
						}
					}
				}
			}

			if isZero(room, walker, col) {
				room.Board[walker][col] = player
				room.CurrentGame.TurnCount++
				fmt.Println("Turn Count", room.CurrentGame.TurnCount)
			} else {
				http.Error(w, "The column has been filled.", 400)
				return
			}

			result := room.CheckBoard(player)

			if result {
				// We have a Winner
				room.DeclareWinner(player)
			} else {

				if room.CurrentGame.TurnCount == 42 {
					// 42 turns have occurred without a Winner; declare a tie.
					room.DeclareTie()
				}
			}

			flip(room)
		} else {
			http.Error(w, "An illegal Turn or illegal play was made", 400)
			return
		}

		// Update the data base with the move
		if room.CurrentGame != nil {
			gameBoardStates := append(duplicate, room.Board)

			// Set the new Board state with the latest move.
			room.CurrentGame.BoardStates = gameBoardStates
			GameDaoInstance().UpdateGame(room.CurrentGame)
			RoomDaoInstance().UpdateRoom(room)

		}

		go room.HandleMove("")

	}
	defer r.Body.Close()
}

func (h *Hub) HandleResignRequest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var resign Resign
	err := decoder.Decode(&resign)

	if err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}

	room := h.hub[resign.Room]
	room.ForceLoss(resign.Player)
}

func (h *Hub) HandleLeaveRequest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var leave Leave
	err := decoder.Decode(&leave)

	if err != nil {
		//log.Println("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}

	room := h.hub[leave.Room]
	playerString := strconv.Itoa(leave.Player)

	if room != nil {
		delete(room.PlayerId, playerString)
		room.Leave(leave.PlayerId)
		room.Blacklist(leave.PlayerId)
		fmt.Println("Removing playerId: ", room.PlayerId[playerString])

		if len(room.PlayerId) == 0 {
			// If the room is empty, remove the room from memory entirely.
			if _, ok := h.hub[leave.Room]; ok {
				delete(h.hub, leave.Room)
				fmt.Println("Deleting Room from memory: ", leave.Room)
			}
		}
	}
	room.HandleMove("Your opponent has left the room.")
}

func (h *Hub) HandleRematchRequest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var rematch Rematch
	err := decoder.Decode(&rematch)

	if err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}

	player := rematch.Player
	roomhash := rematch.Room
	room := h.hub[roomhash]
	playerId := rematch.PlayerId

	if room == nil {
		http.Error(w, "The room does not exist. It's possible your connection was disconnected. Please background the app and start again.", 400)
		return
	}

	if playerId != room.PlayerId[strconv.Itoa(player)] {
		http.Error(w, "You are only allowed to request a rematch on your own behalf.", 400)
		return
	}

	room.RematchRequest[playerId] = true

	playerOne := room.PlayerId["1"]
	playerTwo := room.PlayerId["2"]
	if room.RematchRequest[playerOne] && room.RematchRequest[playerTwo] {
		room.ResetGame()
	}

	// Consider this a state change
	go room.HandleMove("")

	// Return request state
	js, err := json.Marshal(room.RematchRequest)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

	defer r.Body.Close()
}

func (h *Hub) HandleFindAllGamesPlayedRequest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var history GameHistory
	err := decoder.Decode(&history)

	if err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}

	games, err := GameDaoInstance().FindAllCompleteGamesByPlayer(history.PlayerId)
	js, err := json.Marshal(games)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	defer r.Body.Close()
}

func isZero(r *Room, row int, col int) bool {
	return r.Board[row][col] == 0
}

func isTurn(turn int, r *Room, playerId string) bool {
	return r.PlayerId[strconv.Itoa(r.Turn)] == playerId && r.Turn == turn
}

func flip(r *Room) {
	v := r.Turn

	if v == 1 {
		r.Turn = 2
	} else {
		r.Turn = 1
	}
}

/* Constructor */
func NewHub() *Hub {
	hub := new(Hub)
	hub.hub = make(map[string]*Room)
	hub.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return hub
}
