package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

const MaxNumbers = 16

func CalcSquare(m int) (int, error) {
	if rand.Intn(1000) == 555 {
		return 0, fmt.Errorf("got very rare error")
	}
	return m * m, nil
}

func SquareTable(n int) ([]int, error) {
	errChan := make(chan error, 1)

	res := make([]int, n)
	var wg sync.WaitGroup

	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(k int, r []int) {
			defer wg.Done()

			sq, err := CalcSquare(k + 1)
			if err != nil {
				errChan <- err
				return
			}

			r[k] = sq
		}(i, res)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		return nil, err
	}

	return res, nil
}

func PrintTable(results []int) {
	for i := 0; i < len(results); i++ {
		fmt.Printf("\t%d", i+1)
	}
	fmt.Println("")

	for i := 0; i < len(results); i++ {
		fmt.Printf("\t%d", results[i])
	}
	fmt.Println("")
}

func main() {
	table, err := SquareTable(MaxNumbers)
	if err != nil {
		log.Fatalln(err)
	}

	PrintTable(table)
}
