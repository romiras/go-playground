package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/artyom/guesslanguage"
)

func do(reader *bufio.Reader, writer *bufio.Writer) {
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if err = PutLine(writer, line); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		writer.Flush()
	}
}

func detectLanguageRoughly(line string) (string, error) {
	var lang string
	langRuneCounts := make(map[string]uint)
	for _, rune := range line {
		switch {
		case (rune >= 0 && rune <= 0x001F) || (rune >= 0x0020 && rune <= 0x002F) || (rune >= 0x003A && rune <= 0x0040) || (rune >= 0x005B && rune <= 0x0060):
			lang = "!"
		case (rune >= 0x0410 && rune <= 0x044F) || rune == 0x0401 || rune == 0x0451:
			lang = "ru"
		case rune >= 0x0600 && rune <= 0x06FF:
			lang = "ar"
		case rune >= 0x1F600 && rune <= 0x1F64F:
			lang = ":-)"
		default:
			lang = "?"
		}
		_, ok := langRuneCounts[lang]
		if ok {
			langRuneCounts[lang]++
		} else {
			langRuneCounts[lang] = 1
		}
	}
	fmt.Printf("%+v\n", langRuneCounts)
	maxCount := uint(0)
	for l, count := range langRuneCounts {
		if maxCount < count && count >= 3 {
			maxCount = count
			lang = l
		}
	}
	return lang, nil
}

/*
	when 0..0x001F, 0x0020..0x002F, 0x003A..0x0040, 0x005B..0x0060 # C0 controls, punkt
		script = nil
	when 'A'.ord..'Z'.ord, 'a'.ord..'z'.ord
		script = 'Latin'
	when 0x0080..0x00FF, 0x0100..0x017F, 0x0180..0x024F
		script = 'Latin-ext'
	when 0x0410..0x04FE
		script = 'Cyrillic'
	when 0x0391..0x03C9
		script = 'Greek'
	when 0x05D0..0x05EA
		script = 'Hebrew'
	when 0x0600..0x06FF
		script = 'Arabic'
	when 0x0985..0x09F1
		script = 'Bengali'
	when 0x1100..0x11FF, 0xAC00..0xD7A3
		script = 'Hangul'
	when 0x1200..0x137F
		script = 'Ethiopic'
	when 0x30A0..0x30FF
		script = 'Katakana'
	when 0x0E00..0x0E7F
		script = 'Thai'
	when 0x4E00..0x62FF, 0x6300..0x77FF, 0x7800..0x8CFF, 0x8D00..0x9FFF
		script = 'CJK'
	when 0x1F600..0x1F64F
		script = 'Emoticons'
	else
		script = 'Latin?'
		if [' '.ord, '.'.ord, ','.ord, '-'.ord, ':'.ord].include?(c)
			script = nil
		end
	end
*/

func detectLanguage(line string) (string, error) {
	lang, err := detectLanguageRoughly(line)
	if err != nil {
		return "", err
	}

	if lang != "?" {
		return lang, nil
	}

	lang, err = guesslanguage.Guess(line)
	if err != nil {
		return "", err
	}

	// fmt.Println(lang)

	return lang, nil
}

func PutLine(writer *bufio.Writer, line []byte) error {
	lang, err := detectLanguage(string(line))
	if err != nil {
		return err
	}

	_, err = writer.WriteString(fmt.Sprintf("%s\t%s\n", lang, string(line)))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// x, e := detectLanguageRoughly("Монгол Сэтгэгдэл")
	// if e != nil {
	// 	fmt.Println(e.Error())
	// 	os.Exit(1)
	// }
	// fmt.Println(x)
	// return
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	do(reader, writer)
}

/*

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
*/
