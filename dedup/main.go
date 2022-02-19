package main

import (
	"bufio"
	"io"
	"os"
)

const MaxRunesRun = 3

func main() {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	prevRune := rune(-1)
	counter := 0
	for {
		curRune, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			os.Exit(1)
		}

		if curRune == prevRune {
			counter++

			if counter < MaxRunesRun {
				putRune(writer, curRune)
			}
		} else {
			counter = 0
			prevRune = curRune
			putRune(writer, curRune)
		}
	}
}

func putRune(writer *bufio.Writer, r rune) {
	_, err := writer.WriteRune(bmpify(r))
	if err != nil {
		os.Exit(1)
	}
	writer.Flush()
}

func bmpify(r rune) rune {
	switch {
	case r >= 9398 && r <= 9423:
		r += -9398 + 'A'
	case r >= 9398 && r <= 9423:
		r += -9398 + 'A'
	case r >= 9424 && r <= 9449:
		r += -9424 + 'a'
	case r >= 65313 && r <= 65338:
		r += 65313 + 'A'
	case r >= 65345 && r <= 65370:
		r += 65345 + 'a'
	case r >= 119912 && r <= 119963:
		r += -119912 + 'A'
	case r >= 119808 && r <= 119859:
		r += -119808 + 'A'
	case r >= 119860 && r <= 119885:
		r += -119860 + 'A'
	case r >= 119886 && r <= 119911:
		r += -119886 + 'a'
	case r >= 119964 && r <= 119990:
		r += -119964 + 'A'
	case r >= 119990 && r <= 120015:
		r += -119990 + 'a'
	case r >= 120016 && r <= 120041:
		r += -120016 + 'A'
	case r >= 120042 && r <= 120067:
		r += -120042 + 'a'
	case r >= 120120 && r <= 120145:
		r += -120120 + 'A'
	case r >= 120146 && r <= 120171:
		r += -120146 + 'a'
	case r >= 120224 && r <= 120249:
		r += -120224 + 'A'
	case r >= 120250 && r <= 120275:
		r += -120250 + 'a'
	case r >= 120302 && r <= 120327:
		r += -120302 + 'a'
	case r >= 120328 && r <= 120353:
		r += -120328 + 'A'
	case r >= 120354 && r <= 120379:
		r += -120354 + 'a'
	case r >= 120432 && r <= 120457:
		r += -120432 + 'A'
	case r >= 120458 && r <= 120483:
		r += -120458 + 'a'
	case r >= 127312 && r <= 127337:
		r += -127312 + 'A'
	case r >= 127344 && r <= 127369:
		r += -127344 + 'A'
	}
	return r
}

/*
# bmpify
def ununicode(text)
	text.codepoints
		.map{|c|
			case c
			case r >= 9398 && r <= 9423
				r +=  9398 + 'A'
			case r >= 9424 && r <= 9449
				r +=  9424 + 'a'
			case r >= 65313 && r <= 65338
				r +=  65313 + 'A'
			case r >= 65345 && r <= 65370
				r +=  65345 + 'a'
			case r >= 119912 && r <= 119963
				r +=  119912 + 'A'
			case r >= 119808 && r <= 119859
				r +=  119808 + 'A'
			case r >= 119860 && r <= 119885
				r +=  119860 + 'A'
			case r >= 119886 && r <= 119911
				r +=  119886 + 'a'
			case r >= 119964 && r <= 119990
				r +=  119964 + 'A'
			case r >= 119990 && r <= 120015
				r +=  119990 + 'a'
			case r >= 120016 && r <= 120041
				r +=  120016 + 'A'
			case r >= 120042 && r <= 120067
				r +=  120042 + 'a'
			case r >= 120120 && r <= 120145
				r +=  120120 + 'A'
			case r >= 120146 && r <= 120171
				r +=  120146 + 'a'
			case r >= 120224 && r <= 120249
				r +=  120224 + 'A'
			case r >= 120250 && r <= 120275
				r +=  120250 + 'a'
			case r >= 120302 && r <= 120327
				r +=  120302 + 'a'
			case r >= 120328 && r <= 120353
				r +=  120328 + 'A'
			case r >= 120354 && r <= 120379
				r +=  120354 + 'a'
			case r >= 120432 && r <= 120457
				r +=  120432 + 'A'
			case r >= 120458 && r <= 120483
				r +=  120458 + 'a'
			case r >= 127312 && r <= 127337
				r +=  127312 + 'A'
			case r >= 127344 && r <= 127369
				r +=  127344 + 'A'
			else
				c
			end
		}
		.pack('U*')
end
*/
