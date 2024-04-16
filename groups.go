package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
			name := args[0]
			fileName := name + ".kps"
			fileName, err := fixPath(fileName)

			if err != nil {
				panic("invalid name for file!")
			}

			if !doesFileExists(fileName) {
				showError("there is no group with given name", 11)
			}

			f, err := os.OpenFile(fileName, os.O_RDONLY, 0600)
			if err != nil {
				panic("can't open file")
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
