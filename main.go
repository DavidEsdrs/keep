package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"

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
	rootCmd.AddCommand(deleteGroupOrNote())

	// group
	rootCmd.AddCommand(createGroup())

	rootCmd.AddCommand(readFromGroup())
	rootCmd.AddCommand(readGroups())

	rootCmd.PersistentFlags().Bool("desc", false, "Show the notes in decreasing order")

	rootCmd.Execute()
}

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

func deleteGroupOrNote() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [group] [id]",
		Short: "delete a given note within a group - if just the group name is group, the group is deleted",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 2 {
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
			} else if len(args) == 1 {
				groupName := args[0]
				err := DeleteGroup(groupName)
				if err != nil {
					fmt.Printf("unable to delete group %v - error: %v", groupName, err.Error())
					return
				}
				fmt.Printf("group %v deleted", groupName)
			} else {
				panic("invalid args count")
			}
		},
	}
}

func delete() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [id]",
		Short: "removes a given note",
		Args:  cobra.ExactArgs(1),
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

			if err := DeleteNoteById(filename, id); err != nil {
				fmt.Println("unable to delete note with given id")
				return
			}

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
