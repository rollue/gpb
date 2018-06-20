package main

import "fmt"

func conflictPrevent(messages chan Message) {
	dataMap := make(map[string](chan Message))
	finMap := make(map[string](chan bool))

	for msg := range messages {
		//check mgo session is exist

		_, ok := dataMap[msg.Text]
		if ok {
			fmt.Println(fmt.Sprintf("Add data %s", msg.Text))
			dataMap[msg.Text] <- msg
		} else {
			fmt.Println(fmt.Sprintf("Create New Data %s", msg.Text))
			dataMap[msg.Text] = make(chan Message, 100)
			finMap[msg.Text] = make(chan bool)

			go save(dataMap, finMap, msg.Text)
			go finalize(dataMap, finMap, msg.Text)

			dataMap[msg.Text] <- msg
		}

	}
}

func finalize(dataMap map[string](chan Message), finMap map[string](chan bool), msg string) {
	defer delete(finMap, msg)

	for data := range finMap[msg] {
		_, ok := dataMap[msg]
		if ok && data {
			fmt.Println(fmt.Sprintf("Finalize %s", msg))
			delete(dataMap, msg)
		}

		return
	}
}
