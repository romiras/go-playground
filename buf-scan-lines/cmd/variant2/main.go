package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
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
	cmd := exec.Command("ls", "/etc")
	cmd.Stdout = e.wr
	err := cmd.Run()
	return err
}

type BufferedBrokerWriter struct {
	buffer bytes.Buffer
}

func NewBufferedBrokerWriter() *BufferedBrokerWriter {
	return &BufferedBrokerWriter{}
}

func (bw *BufferedBrokerWriter) Write(p []byte) (n int, err error) {
	return bw.buffer.Write(p)
}

func (bw *BufferedBrokerWriter) Read(p []byte) (n int, err error) {
	return bw.buffer.Read(p)
}

func main() {
	// подключаетесь к вашему брокеру сообщений,

	bw := NewBufferedBrokerWriter()

	e := &Executor{}
	e.SetStdout(bw)

	done := make(chan bool)

	err := e.Exec()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		scanner := bufio.NewScanner(bw)
		linesCnt := 0
		for scanner.Scan() {
			fmt.Println(`line:`, scanner.Text()) // для примера (чтобы вы увидели, что данные приходят корректно)
			linesCnt++
			if linesCnt >= 20 {
				// пишите в брокер
				linesCnt = 0
			}
		}
		done <- true
	}()
	<-done

	// time.Sleep(5)
}
