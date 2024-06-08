package main

import (
	"fmt"
	"time"
)

var numChan chan int

func main() {
	input := []int{3, 1, 8, 7, 5, 9, 2}
	numChan = make(chan int)

	go func() {
		fmt.Println("go proc0")
		proc0(numChan)
	}()

	for _, num := range input {
		numChan <- num
	}
}

func proc0(numbers chan int) {
	for num := range numbers {
		time.Sleep(time.Millisecond * 200)
		fmt.Printf("Number %d\n", num)
	}
}
