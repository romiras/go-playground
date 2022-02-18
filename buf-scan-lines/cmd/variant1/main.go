package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

var done chan bool

type Executor struct {
	wr io.Writer
}

func (e *Executor) SetStdout(w io.Writer) {
	e.wr = w
}

func (e *Executor) Exec() error {
	cmd := exec.Command("ticker-demo")
	cmd.Stdout = e.wr
	err := cmd.Run()
	return err
}

func send(lines []string) {
	w := bufio.NewWriter(os.Stdout)
	for i := 0; i < len(lines); i++ {
		w.WriteString(lines[i] + "\n")
	}
	w.Flush()
}

const cnt = 3

func main() {
	// create a pipe
	reader, writer := io.Pipe()

	e := &Executor{}
	e.SetStdout(writer)

	go func() {
		lines := make([]string, 0, cnt)
		scanner := bufio.NewScanner(reader)
		linesCnt := 0
		for scanner.Scan() {
			line := scanner.Text()

			lines = append(lines, line)

			linesCnt++
			if linesCnt >= cnt {
				linesCnt = 0
				send(lines)
				lines = lines[:0] // clear
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}

		if len(lines) > 0 {
			send(lines)
		}
	}()

	err := e.Exec()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Exec finished!")
}
