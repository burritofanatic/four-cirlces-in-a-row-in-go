package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

type GameDAO struct {
	Database *mgo.Database
}

func (s *GameDAO) CreateGame(g *Game) error {
	dbCopy := s.Database.Session.Copy()
	defer dbCopy.Close()
	err := dbCopy.DB(AuthDatabase).C("Game").Insert(g)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func (s *GameDAO) DeleteGame(g *Game) {
	dbCopy := s.Database.Session.Copy()
	defer dbCopy.Close()
	collection := dbCopy.DB(AuthDatabase).C("Game")
	err := collection.Remove(bson.M{"room": g.Room})
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (s *GameDAO) UpdateGame(g *Game) error {
	dbCopy := s.Database.Session.Copy()
	defer dbCopy.Close()
	err := dbCopy.DB(AuthDatabase).C("Game").UpdateId(g.ID, g)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func (s *GameDAO) FindGame(g *Game) (Game, error) {
	dbCopy := s.Database.Session.Copy()
	defer dbCopy.Close()
	var game Game
	err := dbCopy.DB(AuthDatabase).C("Game").FindId(g.ID).One(&game)
	if err != nil {
		fmt.Println(err.Error())
	}
	return game, err
}

func (s *GameDAO) FindAllCompleteGamesByPlayer(d string) ([]Game, error) {
	dbCopy := s.Database.Session.Copy()
	defer dbCopy.Close()
	var games []Game
	err := dbCopy.DB(AuthDatabase).C("Game").Find(bson.M{ "$and": []bson.M{ {"winner": bson.M{"$ne": ""}}, {"turn_count": bson.M{"$gt": 0}}, { "$or": []bson.M{ {"player_one": d}, { "player_two": d}} } } }).Sort("_id").All(&games)
	if err != nil {
		fmt.Println(err.Error())
	}
	return games, err
}

/*A new game is created whenever the last seat is filled if
 there is no incomplete game preexisting between the two players.
In other words, a game without a Winner.*/

// Is there a game where there is no Winner that the user belongs to?
// look for a game where there is no Winner

func (s *GameDAO) FindIncompleteGameForDevice(d string) (Game, error) {
	dbCopy := s.Database.Session.Copy()
	defer dbCopy.Close()
	var game Game
	err := dbCopy.DB(AuthDatabase).C("Game").Find(bson.M{ "$and": []bson.M{ {"winner": ""}, {"turn_count": bson.M{"$gt": 0}}, { "$or": []bson.M{ {"player_one": d}, { "player_two": d}} } } }).One(&game)

	if err != nil {
		fmt.Println(err.Error())
	}
	return game, err
}


var once sync.Once
var sharedInstance *GameDAO

func GameDaoInstance() *GameDAO {
	once.Do(func() {
		sharedInstance = new(GameDAO)
		sharedInstance.Database = NewDB()
	})

	return sharedInstance
}
