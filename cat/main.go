package main

import (
	"bufio"
	"io"
	"os"

	// "github.com/pkg/profile"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	// p := profile.Start(profile.MemProfile)
	// defer p.Stop()
	http.ListenAndServe("localhost:8080", nil)

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	for {
		rune, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			os.Exit(1)
		}

		_, err = writer.WriteRune(rune)
		if err != nil {
			break
			// os.Exit(1)
		}
		writer.Flush()
	}
}
