package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	ErrMalformatedFile = errors.New("info file content is malformated")
)

type NotesInfo struct {
	NotesQuant uint32
	LastUpdate int64 // UNIX timestamp
	CreatedAt  int64
}

func (n NotesInfo) String() string {
	return fmt.Sprintf("%v,%v,%v", info.NotesQuant, info.LastUpdate, info.CreatedAt)
}

func (n *NotesInfo) Add() {
	n.NotesQuant++
}

func (n *NotesInfo) Remove() {
	n.NotesQuant--
}

func (n *NotesInfo) Save() {
	dir, _ := os.UserHomeDir()
	f, err := os.OpenFile(path.Join(dir, ".keep", InfoFilename), os.O_WRONLY, 0700)
	if err != nil {
		panic("something went wrong with the info file! did you delete it?: " + err.Error())
	}
	defer f.Close()

	_, err = f.WriteString(n.String())

	if err != nil {
		panic(fmt.Sprintf("%v%v", "something went wrong while writing into the info file!: ", err.Error()))
	}
}

func doesFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}

func createInfoFile(filename string) error {
	var notesInfo NotesInfo

	targetPath, err := getKeepFilePath()
	if err != nil {
		return err
	}
	targetFile := path.Join(targetPath, filename)

	if doesFileExists(filename) {
		f, err := os.Open(targetFile)
		if err != nil {
			return fmt.Errorf("unable to open info file: %w", err)
		}
		notesInfo, err = parseInfoContent(f)
		if err != nil {
			return fmt.Errorf("unable to parse info file content: %w", err)
		}
	} else {
		f, err := OpenOrCreate(targetFile, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return fmt.Errorf("unable to create info file: %w", err)
		}
		defer f.Close()
		// write basic info
		notesInfo := NotesInfo{}
		content := fmt.Sprintf("%v,%v,%v", notesInfo.NotesQuant, notesInfo.LastUpdate, notesInfo.CreatedAt)
		f.WriteString(content)
	}

	// assign to the global object
	info = &notesInfo

	return nil
}

// return the directory in which the files from keep must be stored
func getKeepFilePath() (string, error) {
	homerDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	keepFolder := path.Join(homerDir, ".keep")
	return keepFolder, err
}

func readFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// parse info file
	notesInfo, err := parseInfoContent(f)

	if err != nil {
		return err
	}

	// assinalar ao objeto global
	info = &notesInfo

	return nil
}

// parse content of given file as keep info
func parseInfoContent(f *os.File) (NotesInfo, error) {
	var (
		notesInfo NotesInfo
		content   string
	)

	s := bufio.NewScanner(f)

	if s.Scan() {
		content = s.Text()
	} else {
		return notesInfo, fmt.Errorf("info file is empty: %w", ErrMalformatedFile)
	}

	segs := strings.Split(content, ",")
	countStr, lastUpdateStr, createdAtStr := segs[0], segs[1], segs[2]

	if count, err := strconv.Atoi(countStr); err != nil {
		return notesInfo, ErrMalformatedFile
	} else {
		notesInfo.NotesQuant = uint32(count)
	}

	if lastUpdate, err := strconv.ParseInt(lastUpdateStr, 10, 64); err != nil {
		return notesInfo, ErrMalformatedFile
	} else {
		notesInfo.LastUpdate = lastUpdate
	}

	if createdAt, err := strconv.ParseInt(createdAtStr, 10, 64); err != nil {
		return notesInfo, ErrMalformatedFile
	} else {
		notesInfo.CreatedAt = createdAt
	}

	return notesInfo, nil
}
