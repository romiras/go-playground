package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const MaxLinesInLogChunk = 10

func run() {
	ticker := time.NewTicker(120 * time.Millisecond)
	done := make(chan bool)

	pr, pw := io.Pipe()

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Fprintf(pw, "Tick at %v\n", t)
			}
		}
	}()

	go scanLogLines(pr)

	time.Sleep(4310 * time.Millisecond)
	ticker.Stop()
	done <- true
}

func main() {
	fmt.Println("Hello, playground")
	run()
	fmt.Println("Ticker stopped")
}

func scanLogLines(reader io.Reader) {
	fmt.Println("scanLogLines")
	lines := make([]string, 0, MaxLinesInLogChunk)
	scanner := bufio.NewScanner(reader)
	linesCnt := uint64(0)
	var mutex = &sync.Mutex{}

	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)

	go func() {
		// for now := range ticker.C {
		// 	fmt.Println("Tick at", now)
		// }
		for {
			select {
			case <-done:
				ticker.Stop()
			case <-ticker.C:
				// flush
				mutex.Lock()
				nLines := len(lines)
				mutex.Unlock()
				if nLines > 0 {
					sendLogLines(lines)
					lines = lines[:0] // clear
				}
				mutex.Lock()
				linesCnt = 0
				mutex.Unlock()
			}
		}
	}()

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)

		mutex.Lock()
		lines = append(lines, line)
		// linesCnt++
		mutex.Unlock()
		atomic.AddUint64(&linesCnt, 1)

		if linesCnt >= MaxLinesInLogChunk {
			linesCnt = 0
			sendLogLines(lines)
			lines = lines[:0] // clear
		}
	}

	fmt.Println("WAIT done")
	done <- true
	fmt.Println("RCV done")

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		log.Fatal("Unable to scan lines. ", err)
	}

	if len(lines) > 0 {
		sendLogLines(lines)
	}
	fmt.Println("/scanLogLines")
}

func sendLogLines(lines []string) {
	log.Printf("Sending %d lines", len(lines))
}
