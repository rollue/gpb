package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type DataLogs struct {
	Text string   `bson:"text"`
	Log  []string `bson:"logs"`
}

func save(dataMap map[string](chan Message), finMap map[string](chan bool), msg string) {

	session, err := mgo.Dial("127.0.0.1")

	if err != nil {
		panic(err)
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("test").C("nouns")

	for value := range dataMap[msg] {
		var results []DataLogs

		err = c.Find(bson.M{"text": msg}).All(&results)

		if len(results) <= 0 {
			c.Insert(&DataLogs{Text: msg, Log: []string{fmt.Sprintf("Add Data from channel %d", value.Channel)}})
		} else {
			pushQuery := bson.M{"$push": bson.M{"logs": fmt.Sprintf("Add Data from channel %d", value.Channel)}}
			who := bson.M{"text": results[0].Text}

			c.Update(who, pushQuery)
		}

		time.Sleep(time.Duration(rand.Int()%3000+2000) * time.Millisecond)

		if len(dataMap[msg]) <= 0 {
			finMap[msg] <- true
			return
		}
	}
}
