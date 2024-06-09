package main

import (
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
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
	return fmt.Sprintf("%v,%v,%v,%v\n", id, now, color, text)
}

func parseTextAsNote(text string) note {
	tokensText := strings.Split(text, ",")

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

// OpenOrCreate opens or creates the file if it doesn't exist.
// Any given folder that doesn't exist will be created recursively
func OpenOrCreate(filename string, flag int, perm fs.FileMode) (*os.File, error) {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("can't create directory: %w", err)
	}
	f, err := os.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("can't open or create file: %w", err)
	}
	return f, nil
}
