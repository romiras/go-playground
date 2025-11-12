// Author: @pav5000 https://qna.habr.com/q/1222600
// Contributor: @romiras
// Machine translation by Google Translate

package main

import (
	"bufio"
	"os"
	"sync"
)

func main() {
	urls := make(chan string)
	go fillChannel(urls)

	// create a group to wait for all workers to be completed
	wg := &sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		// when starting each worker, increase the counter in the group by 1
		wg.Add(1)
		go requestWorker(urls, wg)
	}

	// wait until the counter in the group is equal to 0
	wg.Wait()
}

func requestWorker(channel <-chan string, wg *sync.WaitGroup) {
	// Upon completion of the worker, the counter in the group will be decremented by 1
	defer wg.Done()
	// At the same time we write a message about the completion of the worker
	defer println("Worker stopped")

	// Constantly read new messages from the channel
	// the loop will automatically end when the channel closes and the buffer is empty
	for url := range channel {
		println(url)
	}
}

func fillChannel(channel chan<- string) {
	fileScanner := bufio.NewScanner(os.Stdin)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		channel <- fileScanner.Text()
	}

	// close the channel when the data runs out
	// in Go, it is customary for the channel to be closed only by the one who writes to it
	close(channel)
}
