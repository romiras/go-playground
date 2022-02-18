package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// MaxLines defines number of lines that will be send in one chunk
const MaxLines = 5

var logLinesChan chan []string

func main() {
	var logReader *io.PipeReader
	var logWriter *io.PipeWriter

	logLinesChan = make(chan []string)
	quitChan := make(chan bool)

	logReader, logWriter = io.Pipe()
	go scanLogLines(logReader, quitChan)
	go receiver(quitChan)

	logEmitterA(logWriter)
	// logEmitterB(logWriter)

	_ = logWriter.Close()

	<-quitChan
}

func scanLogLines(reader io.Reader, quitChan chan bool) {
	fmt.Println("!! scanLogLines")

	lines := make([]string, 0, MaxLines)
	scanner := bufio.NewScanner(reader)
	linesCnt := 0
	var mutex = &sync.Mutex{}

	for scanner.Scan() {
		line := scanner.Text()

		fmt.Println("SCAN: " + line)

		mutex.Lock()
		lines = append(lines, line)
		mutex.Unlock()

		linesCnt++
		if linesCnt >= MaxLines {
			linesCnt = 0
			sendLogLines(lines)
			lines = lines[:0] // clear
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		fmt.Println("Unable to scan lines. ", err)
	}

	if len(lines) > 0 {
		sendLogLines(lines)
	}

	close(logLinesChan)

	fmt.Println("!! /scanLogLines")
}

func sendLogLines(lines []string) {
	fmt.Println(fmt.Sprintf(">> log SEND - %v", lines))
	logLinesChan <- lines
}

func receiver(quitChan chan bool) {
	var mutex = &sync.Mutex{}
	for logLines := range logLinesChan {
		mutex.Lock()
		fmt.Printf("\n<< log RCV  - %v\n", logLines)
		mutex.Unlock()
		go publish(logLines)
	}
	quitChan <- true
}

func publish(logLines []string) {
	time.Sleep(5 * time.Millisecond)
	// fmt.Printf("\t@@ PUB (%d)\n", len(logLines))
	// fmt.Println(fmt.Sprintf("<< log RCV  - %v", logLines))
}

var delay = 1 * time.Nanosecond

func logEmitterA(writer io.Writer) {
	for i := 0; i < 12; i++ {
		fmt.Fprintf(writer, "Emitter A %d\n", i+1)
		time.Sleep(delay)
	}

	writer.Write([]byte("1=========================\n"))
}

func logEmitterB(writer io.Writer) {
	for i := 0; i < 12; i++ {
		fmt.Fprintf(writer, "Emitter B %d\n", i+1)
		time.Sleep(delay)
	}

	writer.Write([]byte("2=========================\n"))
}
