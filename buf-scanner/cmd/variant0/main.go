package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

const (
	tickerDuration = 120 * time.Millisecond
	timerDuration  = 1500 * time.Millisecond
)

func main() {
	fmt.Println("Hello, playground")
	run()
}

func run() {
	ticker := time.NewTicker(tickerDuration)
	done := make(chan bool)

	pr, pw := io.Pipe()
	wg := sync.WaitGroup{}
	wg.Add(1)

	log.Println("Ticker started")

	go func() {
		for {
			select {
			case <-done:
				log.Fatal(pw.Close())
				return
			case t := <-ticker.C:
				writeToPipe(pw, t)
			}
		}
	}()

	go func() {
		defer wg.Done()
		scanLogLines(pr)
	}()

	time.Sleep(timerDuration)
	ticker.Stop()
	log.Println("Ticker stopped")

	done <- true

	wg.Wait()
}

func writeToPipe(w *io.PipeWriter, t time.Time) {
	fmt.Fprintf(w, "Tick at %v\n", t)
}

func scanLogLines(reader io.Reader) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		log.Fatal("Unable to scan lines. ", err)
	}
}
