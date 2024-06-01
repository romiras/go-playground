package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	jobsNum    = 5
	workersNum = 2
)

type Job struct {
	id int
}

func worker(id int, jobChan <-chan Job, wg *sync.WaitGroup) {
	for job := range jobChan {
		process(id, job, wg)
	}
}

func process(id int, job Job, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Worker %d, Job start: %d\n", id, job.id)
	time.Sleep(1 * time.Second)
	fmt.Printf("Worker %d, Job end: %d\n", id, job.id)
}

func main() {
	fmt.Println("Start")
	jobChan := make(chan Job, 10)

	var wg sync.WaitGroup

	wg.Add(jobsNum)

	// start the workers
	for i := 0; i < workersNum; i++ {
		go worker(i, jobChan, &wg)
	}

	// enqueue jobs
	for i := 0; i < jobsNum; i++ {
		jobChan <- Job{
			id: i,
		}
	}
	close(jobChan)

	wg.Wait()
	fmt.Println("Done")
}
