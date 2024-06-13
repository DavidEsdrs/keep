package main

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/DavidEsdrs/keep/common"
	"github.com/DavidEsdrs/keep/configs"
	"github.com/DavidEsdrs/keep/notes"
	"github.com/DavidEsdrs/keep/utils"
	"github.com/spf13/cobra"
)

func init() {
	configs.GetDefaultGroupState() // starts default notes file when needed
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

func create() *cobra.Command {
	return &cobra.Command{
		Use:   "[group] [note]",
		Short: "creates a new note",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				err := notes.CreateSingleNote(args[0])
				if err != nil {
					fmt.Println(err)
				}
			} else if len(args) == 2 {
				err := notes.AddNote(args[0], args[1])
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
			_, err := notes.NewNoteFile(groupName, description)
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
				notes, err := notes.ReadAllNotes(groupName + ".kps")
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
				note, err := notes.GetNoteById(groupName, id)
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
			dir, err := utils.GetKeepFilePath()
			if err != nil {
				panic(err)
			}
			f, err := utils.OpenOrCreate(path.Join(dir, common.DEFAULT_KEEP_FILE_PATH), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
			if err != nil {
				panic("can't manage to open file! did you deleted it? error: " + err.Error())
			}
			defer f.Close()

			// TODO: implements --desc flag

			notes, err := notes.ReadAllNotes(common.DEFAULT_KEEP_FILE_PATH)
			if err != nil {
				fmt.Println(err)
				return
			}

			for n := range notes {
				n.Show()
			}

			info := configs.GetDefaultGroupState()

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
				err = notes.DeleteNoteById(groupName, id)
				if err != nil {
					fmt.Printf("unable to delete note %v - error: %v", id, err.Error())
					return
				}
				fmt.Printf("note %v deleted", id)
			} else if len(args) == 1 {
				groupName := args[0]
				err := notes.DeleteGroup(groupName)
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
		Short: "removes a given note from default file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir, err := utils.GetKeepFilePath()
			if err != nil {
				panic(err)
			}
			file, err := utils.OpenOrCreate(path.Join(dir, common.DEFAULT_KEEP_FILE_PATH), os.O_CREATE|os.O_RDWR, 0600)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			id, _ := strconv.ParseInt(args[0], 10, 64)

			if err := notes.DeleteNoteById(common.DEFAULT_KEEP_FILE_PATH, id); err != nil {
				fmt.Println("unable to delete note with given id")
				return
			}

			info := configs.GetDefaultGroupState()

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
			groups, err := notes.GetGroups()
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
