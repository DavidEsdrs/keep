package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidEsdrs/keep/notes"
	"github.com/fatih/color"
)

var (
	Filename     string = "keeps.kps"
	InfoFilename string = "info.kpsinfo"
	NotesFolder  string
	StructSize   int64 = int64(binary.Size(notes.Note{}))
)

var (
	IsProduction bool
)

// global
var (
	Info *NotesInfo
)

type NotesInfo struct {
	NotesQuant  uint32
	LastUpdate  int64 // UNIX timestamp
	CreatedAt   int64
	GroupsCount uint32
}

func (n NotesInfo) String() string {
	return fmt.Sprintf("NotesInfo{%v, %v, %v}", Info.NotesQuant, Info.LastUpdate, Info.CreatedAt)
}

func (n *NotesInfo) Add() {
	n.NotesQuant++
}

func (n *NotesInfo) Remove() {
	n.NotesQuant--
}

func (n *NotesInfo) Save() {
	f, err := os.OpenFile(InfoFilename, os.O_WRONLY, 0600)
	if err != nil {
		panic("something went wrong with the info file! did you deleted it?")
	}
	defer f.Close()

	if err := binary.Write(f, binary.BigEndian, n); err != nil {
		panic(fmt.Sprintf("%v%v", "something went wrong while writing into the info file!: ", err.Error()))
	}
}

func (n *NotesInfo) AddGroup() {
	n.GroupsCount++
}

func createStoreFile() error {
	if !DoesFileExists(Filename) {
		f, err := os.Create(Filename)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}

func createInfoFile() error {
	if DoesFileExists(InfoFilename) {
		return readFile()
	}
	return createFile()
}

func readFile() error {
	f, err := os.Open(InfoFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	// ler as informações binárias
	var notesInfo NotesInfo
	err = binary.Read(f, binary.BigEndian, &notesInfo)

	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	// assinalar ao objeto global
	Info = &notesInfo

	return nil
}

func createFile() error {
	// criar o arquivo
	f, err := os.Create(InfoFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	// escrever informações primárias
	notesInfo := NotesInfo{}

	err = binary.Write(f, binary.BigEndian, &notesInfo)

	// assinalar ao objeto global
	Info = &notesInfo

	return err
}

func RandomColor() color.Attribute {
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

func ShowError(text string, errCode int) {
	errColor := color.New(color.FgHiRed).Add(color.Bold)
	errColor.Println(text)
	os.Exit(errCode)
}

// generates a Note with its date and color
func GenerateNote(text string) notes.Note {
	now := time.Now().UnixMilli()
	color := RandomColor()
	id := Info.NotesQuant + 1

	var textAsBytes [100]byte
	copy(textAsBytes[:], []byte(text))

	return notes.Note{
		Id:        int64(id),
		Text:      textAsBytes,
		CreatedAt: now,
		Color:     int8(color),
	}
}

// generates a Note with its date and color to a group
func GenerateNoteForGroup(text string, groupSize int64) notes.Note {
	now := time.Now().UnixMilli()
	color := RandomColor()
	id := groupSize + 1

	var textAsBytes [100]byte
	copy(textAsBytes[:], []byte(text))

	return notes.Note{
		Id:        int64(id),
		Text:      textAsBytes,
		CreatedAt: now,
		Color:     int8(color),
	}
}

// returns if file with "filename" exists
func DoesFileExists(filename string) bool {
	_, error := os.Stat(filename)
	return !errors.Is(error, os.ErrNotExist)
}

// fixPath fixes a path depending on mode (development or not)
func FixPath(path string) (string, error) {
	if IsProduction {
		ex, err := os.Executable()
		if err != nil {
			return "", err
		}
		ex, err = filepath.EvalSymlinks(ex)
		if err != nil {
			return "", err
		}
		ex = filepath.Dir(ex)
		return fmt.Sprintf("%v\\%v", ex, Filename), nil
	}
	return path, nil // if it is development mode we just create the file in the current directory
}

func init() {
	env := os.Getenv("ENV")

	IsProduction = env == "production"

	ex, _ := os.Executable()
	segs := strings.Split(ex, "\\")

	if IsProduction {
		NotesFolder = strings.Join(segs[:len(segs)-1], "\\")
		ex = filepath.Dir(ex)
		Filename = fmt.Sprintf("%v\\%v", ex, Filename)
		InfoFilename = fmt.Sprintf("%v\\%v", ex, InfoFilename)
	} else {
		NotesFolder, _ = os.Getwd()
	}

	err := createStoreFile()

	if err != nil {
		log.Fatal(err.Error())
	}

	err = createInfoFile()

	if err != nil {
		log.Fatal(err.Error())
	}
}
