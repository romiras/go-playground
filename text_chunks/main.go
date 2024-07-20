package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/textsplitter"
)

const (
	RecursiveCharacterMode = iota
	TokenSplitterMode
	MarkdownTextSplitterMode
)

func main() {
	text, err := readFromStdin()
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}

	splitter := selectSplitter(RecursiveCharacterMode)

	if err := printTextChunks(splitter, text); err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
}

func readFromStdin() (string, error) {
	var text string

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from stdin: %s\n", err.Error())
		return "", err
	}

	return text, nil
}

func selectSplitter(mode int) textsplitter.TextSplitter {
	switch mode {
	case RecursiveCharacterMode:
		return textsplitter.NewRecursiveCharacter()
	case TokenSplitterMode:
		return textsplitter.NewTokenSplitter(textsplitter.WithChunkSize(512), textsplitter.WithChunkOverlap(50))
	case MarkdownTextSplitterMode:
		return textsplitter.NewMarkdownTextSplitter()
	}

	return nil
}

func printTextChunks(splitter textsplitter.TextSplitter, text string) error {
	chunks, err := splitter.SplitText(text)
	if err != nil {
		return err
	}

	for _, chunk := range chunks {
		fmt.Println("---------------")
		fmt.Println(chunk)
	}

	return nil
}
