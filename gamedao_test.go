package main

import "testing"

func DeleteGame (t *testing.T, g *Game) {
	GameDaoInstance().DeleteGame(g)
}

func TestCreateGame(t *testing.T) {

	teardownCase := setupTestCase(t)
	room, _, _ := setupNewRoom(t)

	defer teardownCase(t, room)

	game := room.CreateNewGameWithStartingPositions()
	defer DeleteGame(t, game)

	gameFromDb, err := GameDaoInstance().FindGame(game)

	if !gameFromDb.gameExists() || err != nil {
		t.Error("Game was not saved to the DB")
	}
}

func TestUpdateGame(t *testing.T) {

	teardownCase := setupTestCase(t)
	room, _, _ := setupNewRoom(t)

	defer teardownCase(t, room)

	game := room.CreateNewGameWithStartingPositions()
	defer DeleteGame(t, game)

	game.TurnCount = 42

	GameDaoInstance().UpdateGame(game)
	gameFromDb, _ := GameDaoInstance().FindGame(game)

	if gameFromDb.TurnCount != 42 {
		t.Error("Game was not updated")
	}
}