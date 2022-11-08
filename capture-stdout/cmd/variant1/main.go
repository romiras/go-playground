package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

type Executor struct {
	wr io.WriteCloser
	rd io.Reader
}

func (e *Executor) Exec(name string) error {
	// create a pipe
	e.rd, e.wr = io.Pipe()

	cmd := exec.Command(name)
	cmd.Stdout = e.wr

	wg := sync.WaitGroup{}
	wg.Add(1)
	defer func() {
		if err := e.wr.Close(); err != nil {
			log.Fatal(err)
		}

		wg.Wait() // wait for output of last lines before exit
	}()

	go func() {
		defer wg.Done()
		e.scanLogLines()
	}()

	return cmd.Run()
}

func (e *Executor) scanLogLines() {
	scanner := bufio.NewScanner(e.rd)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		log.Fatal("Unable to scan lines. ", err)
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <program-name>", os.Args[0])
	}
	prog := os.Args[1]

	e := &Executor{}

	if err := e.Exec(prog); err != nil {
		log.Fatal(err)
	}

	log.Println("Exiting.")
}
