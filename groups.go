package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type group struct {
	Id   int64
	Name [10]byte

	// holds the file name in which the notes of  group are stored
	// it is always Name + ".kps"
	Filename    [14]byte
	Description [100]byte

	Size      uint32 // how many items are stored
	Color     int8
	CreatedAt int64
	UpdatedAt int64
}

func (g group) show() {
	fmt.Printf("%v ~ %s\n%v\n%v items", g.Id, string(g.Name[:]), string(g.Description[:]), g.Size)
}

func createGroup() *cobra.Command {
	return &cobra.Command{
		Use:   "group [name] [description]",
		Args:  cobra.ExactArgs(2),
		Short: "creates a group",
		Run: func(cmd *cobra.Command, args []string) {
			var g group

			name := args[0]

			if len(name) > 10 {
				showError("group name too long!", 17)
			}

			var nameBytes [10]byte
			copy(nameBytes[:], name)

			description := args[1]
			if len(name) > 100 {
				showError("group description too long!", 17)
			}

			var descriptionBytes [100]byte
			copy(descriptionBytes[:], []byte(description))

			g.Color = int8(randomColor())
			g.Id = 1
			g.Description = descriptionBytes
			g.Name = nameBytes
			g.Size = 0
			g.CreatedAt = time.Now().UnixMilli()

			fnSlice := nameBytes[:]
			extension := []byte(".kps")
			fullName := bytes.Join([][]byte{
				fnSlice,
				extension,
			}, []byte{})

			copy(g.Filename[:], fullName)

			fileName, err := fixPath(name + ".kps")

			if err != nil {
				panic(err.Error())
			}

			f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				panic("error while opening file " + err.Error())
			}
			defer f.Close()

			err = binary.Write(f, binary.BigEndian, &g)

			if err != nil {
				panic("can't create group file - " + err.Error())
			}
		},
	}
}

// reads information about the group
// in fact, it reads the header of the .kps file correspondent to this group
func descGroup() *cobra.Command {
	return &cobra.Command{
		Use:   "desc [name]",
		Short: "describes the group",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			f, err := getFileForGroup(args[0], os.O_RDONLY, 0600)
			if err != nil {
				showError(err.Error(), 7)
			}
			defer f.Close()

			var g group

			err = binary.Read(f, binary.BigEndian, &g)

			if err != nil {
				panic("can't read file, it may be corrupted")
			}

			g.show()
		},
	}
}

func createNoteInGroup() *cobra.Command {
	return &cobra.Command{
		Use:   "add [group] [note]",
		Short: "add a note to a group",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			file, err := getFileForGroup(args[0], os.O_RDWR, 0666)
			if err != nil {
				showError(err.Error(), 7)
			}
			defer file.Close()

			var g group

			if err := binary.Read(file, binary.BigEndian, &g); err != nil {
				showError(err.Error(), 12)
			}

			// how many to skip: size(group) + size(note) * (how many notes already saved)
			skip := int64(binary.Size(&g)) + (int64(binary.Size(&note{})) * int64(g.Size))

			if _, err := file.Seek(skip, 0); err != nil {
				showError(err.Error(), 12)
			}

			n := generateNoteForGroup(args[1], g)

			if err := binary.Write(file, binary.BigEndian, &n); err != nil {
				panic(err.Error())
			}

			// back to file header
			if _, err := file.Seek(0, 0); err != nil {
				panic(err.Error())
			}

			g.Size++

			if err := binary.Write(file, binary.BigEndian, &g); err != nil {
				panic(err.Error())
			}
		},
	}
}

func getFileForGroup(groupName string, flag int, mode fs.FileMode) (*os.File, error) {
	fileName := groupName + ".kps"
	fileName, err := fixPath(fileName)

	if err != nil {
		return nil, fmt.Errorf("invalid name for file")
	}

	if !doesFileExists(fileName) {
		return nil, fmt.Errorf("there is no group with given name")
	}

	f, err := os.OpenFile(fileName, flag, mode)
	return f, err
}

func readNotesInGroup() *cobra.Command {
	return &cobra.Command{
		Use:     "all [name]",
		Short:   "reads all notes from group with [name]",
		Aliases: []string{"read"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			file, err := getFileForGroup(args[0], os.O_RDONLY, 0600)
			if err != nil {
				showError(err.Error(), 7)
			}
			defer file.Close()

			if _, err := file.Seek(int64(binary.Size(&group{})), 0); err != nil {
				showError(err.Error(), 12)
			}

			var n note

			for {
				err := binary.Read(file, binary.BigEndian, &n)
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
		},
	}
}
