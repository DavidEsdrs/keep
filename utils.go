package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
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
func generateNote(text string) note {
	now := time.Now().UnixMilli()
	color := randomColor()
	id := info.NotesQuant + 1

	var textAsBytes [100]byte
	copy(textAsBytes[:], []byte(text))

	return note{
		Id:        int64(id),
		Text:      textAsBytes,
		CreatedAt: now,
		Color:     int8(color),
	}
}

// generates a note with its date and color to a group
func generateNoteForGroup(text string, g group) note {
	now := time.Now().UnixMilli()
	color := randomColor()
	id := g.Size + 1

	var textAsBytes [100]byte
	copy(textAsBytes[:], []byte(text))

	return note{
		Id:        int64(id),
		Text:      textAsBytes,
		CreatedAt: now,
		Color:     int8(color),
	}
}

// returns if file with "filename" exists
func doesFileExists(filename string) bool {
	_, error := os.Stat(filename)
	return !errors.Is(error, os.ErrNotExist)
}

// fixPath fixes a path depending on mode (development or not)
func fixPath(path string) (string, error) {
	if IsProduction {
		ex, err := os.Executable()
		if err != nil {
			return "", err
		}
		ex, err = filepath.EvalSymlinks(ex) //
		if err != nil {
			return "", err
		}
		ex = filepath.Dir(ex)
		return fmt.Sprintf("%v\\%v", ex, filename), nil
	}
	return path, nil
}
