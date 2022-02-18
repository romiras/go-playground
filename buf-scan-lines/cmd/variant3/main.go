package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"
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

type BufferedBrokerWriter struct {
	buffer     bytes.Buffer
	sendBuffer *bytes.Buffer
	scanner    *bufio.Scanner
}

func NewBufferedBrokerWriter() *BufferedBrokerWriter {
	bw := &BufferedBrokerWriter{}
	bw.scanner = bufio.NewScanner(&bw.buffer)
	bw.sendBuffer = bytes.NewBuffer(make([]byte, 0x10000))
	return bw
}

func (bw *BufferedBrokerWriter) Write(p []byte) (n int, err error) {
	return bw.buffer.Write(p)
}

func (bw *BufferedBrokerWriter) Read(p []byte) (n int, err error) {
	return bw.buffer.Read(p)
}

func (bw *BufferedBrokerWriter) SendBuf() error {
	fmt.Printf("->> %s\n", string(bw.sendBuffer.String()))
	return nil
}

func (bw *BufferedBrokerWriter) ReadLines(cnt uint) error {
	linesCnt := uint(0)
	w := bufio.NewWriter(bw.sendBuffer)

	for bw.scanner.Scan() {
		bytes := bw.scanner.Bytes()
		// err := bw.scanner.Err()
		if /*err != nil ||*/ len(bytes) == 0 {
			// return err
			return nil
		}

		_, err := w.Write(bytes)
		if err != nil {
			fmt.Printf("Write err ->%s<-\n", string(bytes))
			return err
		}
		fmt.Printf("DBG: %s\n", string(bytes))

		linesCnt++
		if linesCnt >= cnt {
			_ = bw.SendBuf()
			return nil
		}
	}

	if err := bw.scanner.Err(); err != nil {
		// fmt.Fprintln(os.Stderr, "reading standard input:", err)
		return err
	}

	return nil
}

func main() {
	bw := NewBufferedBrokerWriter()

	e := &Executor{}
	e.SetStdout(bw)

	ready := make(chan bool)
	done := make(chan bool)

	// timer1 := time.NewTimer(2 * time.Millisecond)
	timer1 := time.NewTimer(2 * time.Second)

	go func() {
		ready <- true
		log.Println("ready - sent")

		log.Println("Exec started!")
		err := e.Exec()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Exec finished!")
	}()

	go func() {
		<-timer1.C // wait a bit before starting

		log.Println("ready rcv: WAIT")
		<-ready
		log.Println("ready rcv: OK")

		err := bw.ReadLines(3)
		if err != nil {
			log.Println(err.Error())
		}

		log.Println("done - sent")
		done <- true
		log.Println("ready done: OK")
	}()

	log.Println("ready done: WAIT")
	<-done
	log.Println("DONE.")
}
