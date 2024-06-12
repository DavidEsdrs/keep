package main

import (
	"bufio"
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
	rootCmd.AddCommand(deleteNote())

	// group
	rootCmd.AddCommand(createGroup())

	rootCmd.AddCommand(deleteGroup())
	rootCmd.AddCommand(readFromGroup())
	rootCmd.AddCommand(readGroups())

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
		Use:     "[group] [note]",
		Aliases: []string{"create", "add"},
		Short:   "creates a new note",
		Args:    cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				err := CreateSingleNote(args[0])
				if err != nil {
					fmt.Println(err)
				}
			} else if len(args) == 2 {
				err := AddNote(args[0], args[1])
				if err != nil {
					fmt.Println(err)
				}
			} else {
				panic("invalid args length")
			}
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
			_, err := NewNoteFile(groupName, description)
			if err != nil {
				panic(err)
			}
			fmt.Printf("group %s created\n", groupName)
		},
	}
}

func readFromGroup() *cobra.Command {
	return &cobra.Command{
		Use:     "read [group] [id]",
		Aliases: []string{},
		Short:   "read notes from group",
		Args:    cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			groupName := args[0]
			if len(args) == 1 {
				notes, err := ReadAllNotes(groupName + ".kps")
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

			// TODO: implements --desc flag

			notes, err := ReadAllNotes(filename)
			if err != nil {
				fmt.Println(err)
				return
			}

			for n := range notes {
				n.Show()
			}

			fmt.Printf("%v notes\n", info.Size)
		},
	}
}

func deleteNote() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [group] [id]",
		Short: "delete a given note within a group",
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

func delete() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [id]",
		Aliases: []string{"forget", "del"},
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

func readGroups() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
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
