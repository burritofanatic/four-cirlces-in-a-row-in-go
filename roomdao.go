package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

type RoomDAO struct {
	Database *mgo.Database
}

func (s *RoomDAO) CreateRoom(r *Room) error {
	dbCopy := s.Database.Session.Copy()
	defer dbCopy.Close()
	err := dbCopy.DB(AuthDatabase).C("Room").Insert(r)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func (s *RoomDAO) UpdateRoom(r *Room) error {
	dbCopy := s.Database.Session.Copy()
	defer dbCopy.Close()
	err := dbCopy.DB(AuthDatabase).C("Room").UpdateId(r.ID, r)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func (s *RoomDAO) FindRoom(n string) (*Room, error) {
	dbCopy := s.Database.Session.Copy()
	defer dbCopy.Close()
	var room *Room
	err := dbCopy.DB(AuthDatabase).C("Room").Find(bson.M{"name": n} ).One(&room)
	if err != nil {
		fmt.Println(err.Error())
	}
	return room, err
}

func (s *RoomDAO) DeleteRoom(n string) {
	dbCopy := s.Database.Session.Copy()
	defer dbCopy.Close()
	collection := dbCopy.DB(AuthDatabase).C("Room")
	err := collection.Remove(bson.M{"name": n})
	if err != nil {
		fmt.Println(err.Error())
	}
}


var onceRoom sync.Once
var roomSharedInstanced *RoomDAO

func RoomDaoInstance() *RoomDAO {
	onceRoom.Do(func() {
		roomSharedInstanced = new(RoomDAO)
		roomSharedInstanced.Database = NewDB()
	})

	return roomSharedInstanced
}