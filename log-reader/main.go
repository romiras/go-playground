package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) != 2 {
		os.Exit(1)
	}
	logFilePath := os.Args[1]

	// fi, err := os.Stat(logFilePath)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Stat: %s, %v\n", fi.Mode().String(), fi)

	cmd := exec.Command("tail", "-f", "-n 20", logFilePath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal("cmd.Start error: " + err.Error())
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("SCAN: %q\n", line)
			if line == "~~END~~" {
				break
			}
		}
		log.Println("scanner exit")
	}()

	if err := cmd.Wait(); err != nil {
		log.Fatal("cmd.Wait error: " + err.Error())
	}

	log.Println("DONE")
}
