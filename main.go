package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

// errors
var (
	ErrUnexEOF = errors.New("keeps: unexpected EOF")
)

// global
var (
	info *NotesInfo
)

type Choose int

const (
	Unknown Choose = iota
	Create
	ReadSingle
	ReadAll
	Delete
)

var (
	filename     string = "keeps.kps"
	infoFilename string = "info.kpsinfo"
	structSize   int64  = int64(binary.Size(note{}))
)

func init() {
	env := os.Getenv("ENV")

	if env == "production" {
		ex, _ := os.Executable()
		ex = filepath.Dir(ex)
		filename = fmt.Sprintf("%v\\%v", ex, filename)
		infoFilename = fmt.Sprintf("%v\\%v", ex, infoFilename)
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

func main() {
	// create is the base command, i.e, when the CLI is called with no
	// subcommands (such as `keep "this a note"`) it is implicity that we want to
	// create a new note
	rootCmd := create()

	// notes commands
	rootCmd.AddCommand(readAll())
	rootCmd.AddCommand(delete())
	rootCmd.AddCommand(readSingle())

	rootCmd.PersistentFlags().Bool("inc", false, "Show the notes in decreasing order")

	rootCmd.Execute()
}

func create() *cobra.Command {
	return &cobra.Command{
		Use:     "[input]",
		Aliases: []string{"create", "add"},
		Short:   "creates a new note",
		Args:    cobra.RangeArgs(1, 10),
		Run: func(cmd *cobra.Command, args []string) {
			f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			for _, note := range args {
				if len(note) <= 100 {
					n := generateNote(note)
					err := binary.Write(f, binary.BigEndian, &n)
					if err != nil {
						panic(err.Error())
					}
					info.Add()
				} else {
					showError("the length of the note is bigger than allowed!", 10)
				}
			}

			info.Save()
		},
	}
}

func readAll() *cobra.Command {
	return &cobra.Command{
		Use:     "all",
		Aliases: []string{"remind", "get"},
		Short:   "remind you all notes",
		Run: func(cmd *cobra.Command, args []string) {
			f, err := os.Open(filename)
			if err != nil {
				panic("can't manage to open file! did you deleted it? error: " + err.Error())
			}
			defer f.Close()

			isIncreasing, _ := cmd.Flags().GetBool("inc")

			if isIncreasing {
				showIncreasingOrder(f)
			} else {
				showDecreasingOrder(f)
			}

			fmt.Printf("%v notes\n", info.NotesQuant)
		},
	}
}

func showIncreasingOrder(f *os.File) {
	var n note

	for {
		err := binary.Read(f, binary.BigEndian, &n)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				panic("keep: error while reading binary file")
			}
		}
		if n.Id != 0 {
			n.show()
		}
	}
}

func showDecreasingOrder(f *os.File) {
	var notes []note
	var n note

	for {
		err := binary.Read(f, binary.BigEndian, &n)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				panic("keep: error while reading binary file. error -" + err.Error())
			}
		}
		notes = append(notes, n)
	}

	for i := len(notes) - 1; i >= 0; i-- {
		if notes[i].Id != 0 {
			note.show(notes[i])
		}
	}
}

func delete() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [id]",
		Aliases: []string{"forget"},
		Short:   "deletes a given note",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			file, err := os.OpenFile(filename, os.O_RDWR, 0666)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			id, _ := strconv.ParseInt(args[0], 10, 64)

			positionToRemove := (id - 1) * structSize

			_, err = file.Seek(positionToRemove, io.SeekStart)

			if err != nil {
				showError("can't remove. is it a valid ID?", 11)
			}

			err = binary.Write(file, binary.BigEndian, &note{}) // we override the line with a empty struct

			if err != nil {
				panic(err)
			}

			info.Remove()
			info.Save()
		},
	}
}

func readSingle() *cobra.Command {
	return &cobra.Command{
		Use:   "read [id]",
		Short: "return a single note",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			file, err := os.OpenFile(filename, os.O_RDWR, 0644)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			id, _ := strconv.ParseInt(args[0], 10, 64)

			if _, err := file.Seek((id-1)*structSize, io.SeekStart); err != nil {
				panic("invalid seeking! error - " + err.Error())
			}

			var n note

			if err := binary.Read(file, binary.BigEndian, &n); err != nil {
				panic("keep: error while reading binary file")
			}

			if n.Id == id { // assert
				n.show()
			}
		},
	}
}
