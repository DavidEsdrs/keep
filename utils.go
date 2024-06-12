package main

import (
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

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

func ExtractExtension(filename string) string {
	segs := strings.Split(filename, ".")
	return segs[len(segs)-1]
}
