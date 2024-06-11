package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path"
)

var (
	ErrMalformatedFile = errors.New("info file content is malformated")
)

type NotesInfo struct {
	Title       [20]rune
	Description [200]rune
	Size        uint32
	SizeAlltime uint32
	LastUpdate  int64 // UNIX timestamp
	CreatedAt   int64
}

func (n NotesInfo) String() string {
	return fmt.Sprintf("%v,%v,%v", n.Size, n.LastUpdate, n.CreatedAt)
}

func (n *NotesInfo) Add() {
	n.SizeAlltime++
	n.Size++
}

func (n *NotesInfo) Remove() {
	n.Size--
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

func createInfoFile(filename string) (*NotesInfo, error) {
	var notesInfo NotesInfo

	targetPath, err := GetKeepFilePath()
	if err != nil {
		return &notesInfo, err
	}
	targetFile := path.Join(targetPath, filename)

	if doesFileExists(targetFile) {
		f, err := os.OpenFile(targetFile, os.O_RDWR, 0)
		if err != nil {
			return &notesInfo, fmt.Errorf("unable to open info file: %w", err)
		}
		notesInfo, err = parseInfoContent(f)
		if err != nil {
			return &notesInfo, fmt.Errorf("unable to parse info file content: %w", err)
		}
	} else {
		f, err := OpenOrCreate(targetFile, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return &notesInfo, fmt.Errorf("unable to create info file: %w", err)
		}
		defer f.Close()
		// write basic info
		notesInfo := NotesInfo{}
		err = binary.Write(f, binary.BigEndian, &notesInfo)
		if err != nil {
			return nil, fmt.Errorf("unable to write data into info file: %w", err)
		}
	}

	// assign to the global object
	info = &notesInfo

	return &notesInfo, nil
}

// return the directory in which the files from keep must be stored
func GetKeepFilePath() (string, error) {
	homerDir, err := os.UserHomeDir()
	keepFolder := path.Join(homerDir, ".keep")
	return keepFolder, err
}

// parse content of given file as keep info
func parseInfoContent(f *os.File) (NotesInfo, error) {
	var notesInfo NotesInfo

	if err := binary.Read(f, binary.BigEndian, &notesInfo); err != nil {
		return notesInfo, nil
	}

	return notesInfo, nil
}
