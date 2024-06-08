package main

import (
	"fmt"
	"time"
)

var numChan chan int
var labelChan chan string

func main() {
	input := []int{3, 1, 8, 0, 7, 5, 9, 2}
	numChan = make(chan int)
	labelChan = make(chan string)

	go func() {
		labelsEmitter(numChan, labelChan)
		close(numChan)
	}()

	go func() {
		labelsReceiver(labelChan)
		close(labelChan)
	}()

	for _, num := range input {
		fmt.Println(num)
		numChan <- num
	}

	time.Sleep(time.Millisecond * 1000)
}

func labelsEmitter(numbers chan int, outChan chan string) {
	fmt.Println("go labelsEmitter")
	for num := range numbers {
		time.Sleep(time.Millisecond * 250)
		str := fmt.Sprintf("Number %d", num)
		outChan <- str
	}
}

func labelsReceiver(labelChanRcv chan string) {
	fmt.Println("go labelsReceiver")
	for label := range labelChanRcv {
		fmt.Println(label)
	}
}
