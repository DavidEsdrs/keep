package configs

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/DavidEsdrs/keep/common"
	"github.com/DavidEsdrs/keep/utils"
)

var info *NotesInfo // interface for default notes file

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
	f, err := os.OpenFile(path.Join(dir, ".keep", common.INFO_FILE_PATH), os.O_WRONLY, 0700)
	if err != nil {
		panic("something went wrong with the info file! did you delete it?: " + err.Error())
	}
	defer f.Close()

	_, err = f.WriteString(n.String())
	if err != nil {
		panic(fmt.Sprintf("%v%v", "something went wrong while writing into the info file!: ", err.Error()))
	}
}

// get singleton
func GetDefaultGroupState() *NotesInfo {
	if info == nil {
		inf, err := createInfoFile(common.INFO_FILE_PATH)
		if err != nil {
			panic("unable to manage default note file")
		}
		info = inf
	}
	return info
}

func createInfoFile(filename string) (*NotesInfo, error) {
	var notesInfo NotesInfo

	targetPath, err := utils.GetKeepFilePath()
	if err != nil {
		return &notesInfo, err
	}
	targetFile := path.Join(targetPath, filename)

	if utils.DoesFileExists(targetFile) {
		f, err := os.OpenFile(targetFile, os.O_RDWR, 0)
		if err != nil {
			return &notesInfo, fmt.Errorf("unable to open info file: %w", err)
		}
		notesInfo, err = parseInfoContent(f)
		if err != nil {
			return &notesInfo, fmt.Errorf("unable to parse info file content: %w", err)
		}
	} else {
		f, err := utils.OpenOrCreate(targetFile, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return &notesInfo, fmt.Errorf("unable to create info file: %w", err)
		}
		defer f.Close()
		// write basic info

		var (
			title       [20]rune
			description [200]rune
		)

		copy(title[:], []rune("default"))
		copy(description[:], []rune("this is the default note group - notes with no group given will be stored here"))

		now := time.Now().UnixMilli()

		notesInfo := NotesInfo{
			Title:       title,
			Description: description,
			Size:        0,
			SizeAlltime: 0,
			LastUpdate:  now,
			CreatedAt:   now,
		}
		err = binary.Write(f, binary.BigEndian, &notesInfo)
		if err != nil {
			return nil, fmt.Errorf("unable to write data into info file: %w", err)
		}
	}

	// assign to the global object
	info = &notesInfo

	return &notesInfo, nil
}

func parseInfoContent(f *os.File) (NotesInfo, error) {
	var notesInfo NotesInfo

	if err := binary.Read(f, binary.BigEndian, &notesInfo); err != nil {
		return notesInfo, nil
	}

	return notesInfo, nil
}
