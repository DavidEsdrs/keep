package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

func randomColor() color.Attribute {
	randomN := rand.Intn(5)

	switch randomN {
	case 1:
		return color.FgRed
	case 2:
		return color.FgGreen
	case 3:
		return color.FgBlue
	case 4:
		return color.FgYellow
	case 5:
		return color.FgBlack
	case 6:
		return color.FgWhite
	case 7:
		return color.FgCyan
	case 8:
		return color.FgMagenta
	default:
		return color.FgCyan
	}

}

func showError(text string, errCode int) {
	errColor := color.New(color.FgHiRed).Add(color.Bold)
	errColor.Println(text)
	os.Exit(errCode)
}

// generates a note with its date and color
func generateNote(text string) string {
	now := time.Now().UnixMilli()
	color := randomColor()
	id := info.NotesQuant + 1
	return fmt.Sprintf("[%v]{%v}(%v)`%v`\n", id, now, color, text)
}

func parseTextAsNote(text string) note {
	result, err := processor.Tokenize(text)
	if err != nil {
		showError(err.Error(), 9)
	}
	tokens := result.Tokens()

	var tokensText [4]string

	for i, t := range tokens {
		tokensText[i] = t.InnerText
	}

	id, _ := strconv.ParseInt(tokensText[0], 10, 64)

	date, err := strconv.ParseInt(tokensText[1], 10, 64)

	if err != nil {
		showError("invalid date", 8)
	}

	colorNumber, _ := strconv.Atoi(tokensText[2])
	colorAttribute := color.Attribute(colorNumber)

	if err != nil {
		showError("invalid color", 8)
	}

	innerText := tokensText[3]

	return note{
		id:        id,
		createdAt: date,
		text:      innerText,
		color:     colorAttribute,
	}
}

func colorFromString(c string) (color.Attribute, error) {
	switch strings.ToLower(c) {
	case "red":
		return color.FgRed, nil
	case "green":
		return color.FgGreen, nil
	case "blue":
		return color.FgBlue, nil
	case "yellow":
		return color.FgYellow, nil
	case "black":
		return color.FgBlack, nil
	case "white":
		return color.FgWhite, nil
	case "cyan":
		return color.FgCyan, nil
	case "magenta":
		return color.FgMagenta, nil
	default:
		return color.Attribute(0), fmt.Errorf("cor não reconhecida")
	}
}
