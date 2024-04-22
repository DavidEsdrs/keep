package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/DavidEsdrs/keep/common"
	"github.com/DavidEsdrs/keep/groups"
	"github.com/DavidEsdrs/keep/notes"
	"github.com/spf13/cobra"
)

// errors
var (
	ErrUnexEOF = errors.New("keeps: unexpected EOF")
)

type Choose int

const (
	Unknown Choose = iota
	Create
	ReadSingle
	ReadAll
	Delete
)

func main() {
	// create is the base command, i.e, when the CLI is called with no
	// subcommands (such as `keep "this a note"`) it is implicity that we want to
	// create a new note
	rootCmd := create()

	// notes commands
	rootCmd.AddCommand(readAll())
	rootCmd.AddCommand(delete())
	rootCmd.AddCommand(readSingle())

	// groups commands
	group := groups.CreateGroup()

	rootCmd.AddCommand(group)
	group.AddCommand(groups.DescGroup())
	group.AddCommand(groups.CreateNoteInGroup())
	group.AddCommand(groups.ReadNotesInGroup())
	group.AddCommand(groups.ReadNoteFromGroup())
	group.AddCommand(groups.GetGroups())

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
			f, err := os.OpenFile(common.Filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			for _, note := range args {
				if len(note) <= 100 {
					n := common.GenerateNote(note)
					err := binary.Write(f, binary.BigEndian, &n)
					if err != nil {
						panic(err.Error())
					}
					common.Info.Add()
				} else {
					common.ShowError("the length of the note is bigger than allowed!", 10)
				}
			}

			common.Info.Save()
		},
	}
}

func readAll() *cobra.Command {
	return &cobra.Command{
		Use:     "all",
		Aliases: []string{"remind", "get"},
		Short:   "remind you all notes",
		Run: func(cmd *cobra.Command, args []string) {
			f, err := os.Open(common.Filename)
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

			fmt.Printf("%v notes\n", common.Info.NotesQuant)
		},
	}
}

func showIncreasingOrder(f *os.File) {
	var n notes.Note

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
			n.Show()
		}
	}
}

func showDecreasingOrder(f *os.File) {
	var ns []notes.Note
	var n notes.Note

	for {
		err := binary.Read(f, binary.BigEndian, &n)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				panic("keep: error while reading binary file. error -" + err.Error())
			}
		}
		ns = append(ns, n)
	}

	for i := len(ns) - 1; i >= 0; i-- {
		if ns[i].Id != 0 {
			n.Show()
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
			file, err := os.OpenFile(common.Filename, os.O_RDWR, 0666)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			id, _ := strconv.ParseInt(args[0], 10, 64)

			positionToRemove := (id - 1) * common.StructSize

			_, err = file.Seek(positionToRemove, io.SeekStart)

			if err != nil {
				common.ShowError("can't remove. is it a valid ID?", 11)
			}

			err = binary.Write(file, binary.BigEndian, &notes.Note{}) // we override the line with a empty struct

			if err != nil {
				panic(err)
			}

			common.Info.Remove()
			common.Info.Save()
		},
	}
}

func readSingle() *cobra.Command {
	return &cobra.Command{
		Use:   "read [id]",
		Short: "return a single note",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			file, err := os.OpenFile(common.Filename, os.O_RDWR, 0644)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			id, _ := strconv.ParseInt(args[0], 10, 64)

			if _, err := file.Seek((id-1)*common.StructSize, io.SeekStart); err != nil {
				panic("invalid seeking! error - " + err.Error())
			}

			var n notes.Note

			if err := binary.Read(file, binary.BigEndian, &n); err != nil {
				panic("keep: error while reading binary file")
			}

			if n.Id == id { // assert
				n.Show()
			}
		},
	}
}
