package main

import "testing"

func TestEmptyBoard(t *testing.T) {
	room := NewRoom("test")
	resultOne := room.CheckBoard(1)

	if resultOne != false {
		t.Error("Empty board should have no winners")
	}

	resultTwo := room.CheckBoard(2)
	if resultTwo != false {
		t.Error("Empty board should have no winners")
	}
}

func TestHorizontalWin(t *testing.T) {
	room := NewRoom("test")
	room.Board = [][]int{
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 2, 0, 0, 0} ,
		{0, 0, 0, 2, 0, 0, 0} ,
		{0, 0, 0, 2, 0, 0, 0} ,
		{0, 0, 2, 0, 0, 0, 0} ,
		{1, 1, 1, 1, 0, 0, 0} ,
	}

	resultOne := room.CheckBoard(1)
	if resultOne != true {
		t.Error("Player One should have won")
	}
}

func TestVerticalWin(t *testing.T) {
	room := NewRoom("test")
	room.Board = [][]int{
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 2, 0, 0, 0} ,
		{0, 0, 0, 2, 0, 0, 0} ,
		{0, 0, 0, 2, 0, 0, 0} ,
		{0, 0, 2, 2, 0, 0, 0} ,
		{1, 1, 2, 1, 0, 0, 0} ,
	}

	resultOne := room.CheckBoard(2)
	if resultOne != true {
		t.Error("Player Two should have won")
	}
}

func TestAscendingDiagonalWin(t *testing.T) {
	room := NewRoom("test")
	room.Board = [][]int{
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 2, 0, 0, 0} ,
		{0, 0, 0, 1, 0, 0, 0} ,
		{0, 0, 1, 2, 0, 0, 0} ,
		{0, 1, 2, 2, 0, 0, 0} ,
		{1, 1, 2, 1, 0, 0, 0} ,
	}

	resultOne := room.CheckBoard(1)
	if resultOne != true {
		t.Error("Player One should have won")
	}
}

func TestDescendingDiagonalWin(t *testing.T) {
	room := NewRoom("test")
	room.Board = [][]int{
		{0, 0, 0, 0, 0, 0, 0} ,
		{0, 0, 0, 2, 0, 0, 0} ,
		{0, 2, 0, 1, 0, 0, 0} ,
		{0, 0, 2, 2, 0, 0, 0} ,
		{0, 1, 2, 2, 0, 0, 0} ,
		{1, 1, 2, 1, 2, 0, 0} ,
	}

	resultOne := room.CheckBoard(2)
	if resultOne != true {
		t.Error("Player Two should have won")
	}
}