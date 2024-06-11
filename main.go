package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	ErrUnexEOF = errors.New("keeps: unexpected EOF")
)

var (
	info         *NotesInfo
	InfoFilename string = "info.kps"
)

func init() {
	_, err := createInfoFile(InfoFilename)

	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	// create is the base command, i.e, when the CLI is called with no
	// subcommands (such as `keep "this a note"`) it is implicity that we want to
	// create a new note
	rootCmd := create()

	rootCmd.AddCommand(readAll())
	rootCmd.AddCommand(delete())
	rootCmd.AddCommand(readSingle())
	rootCmd.AddCommand(deleteNote())

	// group
	groupCmd := createGroup()
	rootCmd.AddCommand(groupCmd)

	groupCmd.AddCommand(addNoteInGroup())
	groupCmd.AddCommand(readFromGroup())
	groupCmd.AddCommand(deleteGroup())
	groupCmd.AddCommand(readGroups())

	rootCmd.PersistentFlags().Bool("desc", false, "Show the notes in decreasing order")

	rootCmd.Execute()
}

type note struct {
	id        int64
	text      string
	color     color.Attribute
	createdAt int64
}

func (n note) show() {
	c := color.New(n.color).Add(color.Bold)

	t := time.Unix(n.createdAt/1000, 0)

	blue := color.New(color.BgHiBlue).Add(color.Bold)

	blue.DisableColor()

	blue.Print(fmt.Sprint(n.id) + " ~ ")

	blue.EnableColor()

	now := time.Now()

	if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() {
		blue.Printf(" %v ", t.Local().Format(time.Kitchen))
	} else {
		blue.Printf(" %v ", t.Local().Format("01/02/2006"))
	}

	blue.DisableColor()
	blue.Print(" - ")
	blue.EnableColor()
	c.Print(n.text)
	c.Println()
}

type Choose int

const (
	Unknown Choose = iota
	Create
	ReadSingle
	ReadAll
	Delete
)

const filename string = "keeps.txt"

func create() *cobra.Command {
	return &cobra.Command{
		Use:     "[input]",
		Aliases: []string{"create", "add"},
		Short:   "creates a new note",
		Args:    cobra.RangeArgs(1, 10),
		Run: func(cmd *cobra.Command, args []string) {
			dir, err := GetKeepFilePath()
			if err != nil {
				panic(err)
			}
			f, err := OpenOrCreate(path.Join(dir, filename), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			w := bufio.NewWriter(f)

			for _, note := range args {
				if len(note) <= 100 {
					note = generateNote(note)
					n, err := w.WriteString(note)
					if err != nil {
						panic("error while writing: " + err.Error())
					}
					if n == 0 {
						fmt.Println("no data written with no error")
					}
					info.Add()
				} else {
					showError("the length of the note is bigger than allowed!", 10)
				}
			}

			err = w.Flush()
			if err != nil {
				panic("error while flushing! " + err.Error())
			}
			info.Save()
		},
	}
}

func createGroup() *cobra.Command {
	return &cobra.Command{
		Use:     "group [name] [desc]",
		Aliases: []string{},
		Short:   "creates a new note group",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			groupName := args[0]
			description := args[1]
			h, err := NewNoteFile(groupName, description)
			if err != nil {
				panic(err)
			}
			fmt.Printf("group %s created\n", string(h.Title[:]))
		},
	}
}

func addNoteInGroup() *cobra.Command {
	return &cobra.Command{
		Use:     "add [group] [message]",
		Aliases: []string{},
		Short:   "adds note in group",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			h, err := GetGroupHeader(args[0])
			if err != nil {
				panic(err)
			}
			n := NewNote(int64(h.SizeAlltime+1), args[1], randomColor(), time.Now().UnixMilli())
			err = AddNote(args[0], &n)
			if err != nil {
				fmt.Println(err)
			}
		},
	}
}

func readFromGroup() *cobra.Command {
	return &cobra.Command{
		Use:     "get [group] [id]",
		Aliases: []string{},
		Short:   "read notes from group",
		Args:    cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			groupName := args[0]
			if len(args) == 1 {
				notes, err := ReadAllNotes(groupName)
				if err != nil {
					panic(err)
				}
				for n := range notes {
					n.Show()
				}
			} else if len(args) == 2 {
				id, err := strconv.ParseInt(args[1], 10, 64)
				if err != nil {
					fmt.Println("id is not a valid number")
				}
				note, err := GetNoteById(groupName, id)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				note.Show()
			}
		},
	}
}

func readAll() *cobra.Command {
	return &cobra.Command{
		Use:     "all",
		Aliases: []string{"remind", "get"},
		Short:   "remind you all notes",
		Run: func(cmd *cobra.Command, args []string) {
			dir, err := GetKeepFilePath()
			if err != nil {
				panic(err)
			}
			f, err := OpenOrCreate(path.Join(dir, filename), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
			if err != nil {
				panic("can't manage to open file! did you deleted it? error: " + err.Error())
			}
			defer f.Close()

			s := bufio.NewScanner(f)

			isDecreasing, _ := cmd.Flags().GetBool("desc")

			if isDecreasing {
				showDecreasingOrder(s)
			} else {
				showIncreasingOrder(s)
			}

			fmt.Printf("%v notes\n", info.Size)
		},
	}
}

func deleteNote() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [group] [id]",
		Aliases: []string{"remind", "get"},
		Short:   "delete a given note within a group",
		Run: func(cmd *cobra.Command, args []string) {
			groupName := args[0]
			id, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				fmt.Println("invalid id given")
				return
			}
			err = DeleteNoteById(groupName, id)
			if err != nil {
				fmt.Printf("unable to delete note %v - error: %v", id, err.Error())
				return
			}
			fmt.Printf("note %v deleted", id)
		},
	}
}

func deleteGroup() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [group]",
		Aliases: []string{},
		Short:   "delete a group",
		Run: func(cmd *cobra.Command, args []string) {
			groupName := args[0]
			err := DeleteGroup(groupName)
			if err != nil {
				fmt.Printf("unable to delete group %v - error: %v", groupName, err.Error())
				return
			}
			fmt.Printf("group %v deleted", groupName)
		},
	}
}

func showIncreasingOrder(s *bufio.Scanner) {
	for s.Scan() {
		text := s.Text()
		note := parseTextAsNote(text)
		note.show()
	}
}

func showDecreasingOrder(s *bufio.Scanner) {
	var notes []note

	for s.Scan() {
		text := s.Text()
		note := parseTextAsNote(text)
		notes = append(notes, note)
	}

	for i := len(notes) - 1; i >= 0; i-- {
		note.show(notes[i])
	}
}

func delete() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [id]",
		Aliases: []string{"forget"},
		Short:   "deletes a given note",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir, err := GetKeepFilePath()
			if err != nil {
				panic(err)
			}
			file, err := OpenOrCreate(path.Join(dir, filename), os.O_CREATE|os.O_RDWR, 0600)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			id, _ := strconv.ParseInt(args[0], 10, 64)

			s := bufio.NewScanner(file)
			w := bufio.NewWriter(file)

			var lines []string

			for s.Scan() {
				line := s.Text()

				note := parseTextAsNote(line)

				if note.id != id {
					lines = append(lines, line)
				}
			}

			err = file.Truncate(0) // cleans the file

			if err != nil {
				panic("keeps: error while truncating file! error - " + err.Error())
			}

			_, err = file.Seek(0, 0)

			if err != nil {
				panic("keeps: error while seeking file! error - " + err.Error())
			}

			for _, l := range lines {
				w.WriteString(fmt.Sprintf("%v\n", l))
			}

			w.Flush()

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
			dir, err := GetKeepFilePath()
			if err != nil {
				panic(err)
			}
			file, err := OpenOrCreate(path.Join(dir, filename), os.O_CREATE|os.O_RDONLY, 0600)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			id, _ := strconv.ParseInt(args[0], 10, 64)

			s := bufio.NewScanner(file)
			found := false

			for s.Scan() {
				line := s.Text()

				note := parseTextAsNote(line)

				if note.id == id {
					note.show()
					found = true
					break
				}
			}

			if !found {
				fmt.Println("not found note with given id")
			}
		},
	}
}

func readGroups() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "get all groups created",
		Run: func(cmd *cobra.Command, args []string) {
			groups, err := GetGroups()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			for _, g := range groups {
				g.Show()
			}
		},
	}
}
