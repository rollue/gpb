package main

import "sync"

var wg sync.WaitGroup

func main() {
	ch := make(chan Message, 100)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go receiver(i, ch)
	}

	go conflictPrevent(ch)

	wg.Wait()
}
