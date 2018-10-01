package main

import (
	"testing"
)

func setupTestCase(t *testing.T) func(t *testing.T, r *Room) {

	return func(t *testing.T, r *Room) {
		RoomDaoInstance().DeleteRoom(r.Name)
	}
}

func setupNewRoom(t *testing.T) (*Room, Game, *Hub) {
	h := NewHub()
	room, game := h.FindSeat("test")
	return room, game, h
}

func TestNewRoom(t *testing.T) {

	teardownCase := setupTestCase(t)
	room, game, _ := setupNewRoom(t)

	defer teardownCase(t, room)

	if room == nil {
		t.Error("Room is nil")
	}

	if game.ID != "" {
		t.Error("No game should have been created")
	}

	roomFromDb, err := RoomDaoInstance().FindRoom(room.Name)
	if roomFromDb == nil || err != nil {
		t.Error("Room was not persisted in the Database")
	}
}