package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"bufio"
	"log"
)

// ANSI escape codes for highlighting
const (
	brightRedStart = "\033[91m"
	resetColor     = "\033[0m"
)

var errorRegex *regexp.Regexp

func init() {
	var err error
	errorRegex, err = regexp.Compile("(?i)error")
	if err != nil {
		log.Fatal("Failed to compile regex:", err)
	}
}

func handleError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func highlightLine(lineBytes []byte, regex *regexp.Regexp) ([]byte, error) {
	matches := regex.FindAllSubmatchIndex(lineBytes, -1)

	if len(matches) == 0 {
		return append(lineBytes, '\n'), nil
	}

	var builder strings.Builder
	lastIndex := 0

	for _, match := range matches {
		// Add text before match
		builder.Write(lineBytes[lastIndex:match[0]])

		// Add highlighted match
		builder.WriteString(brightRedStart)
		builder.Write(lineBytes[match[0]:match[1]])
		builder.WriteString(resetColor)

		lastIndex = match[1]
	}

	// Add remaining text
	builder.Write(lineBytes[lastIndex:])
	builder.WriteByte('\n')

	return []byte(builder.String()), nil
}

func processInput() error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 1024), 4096)

	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		highlightedLine, err := highlightLine(lineBytes, errorRegex)
		if err != nil {
			return err
		}

		fmt.Print(string(highlightedLine))
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := processInput(); err != nil {
		handleError(err)
	}
}
