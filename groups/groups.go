package groups

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/DavidEsdrs/keep/common"
	"github.com/DavidEsdrs/keep/notes"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// 8 + 10 + 14 + 100 + 4 + 1 + 8 + 8 = 153 bytes
type Group struct {
	Id   int64
	Name [10]byte

	// holds the file name in which the notes of  Group are stored
	// it is always Name + ".kps"
	Filename    [14]byte
	Description [100]byte

	Size      uint32 // how many items are stored
	Color     int8
	CreatedAt int64
	UpdatedAt int64
}

func (g Group) Short() {
	c := color.New(color.Attribute(g.Color))
	c.Add(color.Bold)
	c.DisableColor()
	c.Print(fmt.Sprint(g.Id) + " ~ ")
	c.EnableColor()
	c.Printf("%s", g.Name)
	c.Println()
}

func (g Group) show() {
	fmt.Printf("%v ~ %s\n%v\n%v items\n", g.Id, string(g.Name[:]), string(g.Description[:]), g.Size)
}

func CreateGroup() *cobra.Command {
	return &cobra.Command{
		Use:   "group [name] [description]",
		Args:  cobra.ExactArgs(2),
		Short: "creates a Group",
		Run: func(Cmd *cobra.Command, args []string) {
			var g Group

			name := args[0]

			if len(name) > 10 {
				common.ShowError("Group name too long!", 17)
			}

			var nameBytes [10]byte
			copy(nameBytes[:], name)

			description := args[1]
			if len(name) > 100 {
				common.ShowError("Group description too long!", 17)
			}

			var descriptionBytes [100]byte
			copy(descriptionBytes[:], []byte(description))

			newId := int64(common.Info.GroupsCount) + 1

			g.Color = int8(common.RandomColor())
			g.Id = newId
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

			fileName, err := common.FixPath(name + ".kps")

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
				panic("can't create Group file - " + err.Error())
			}

			common.Info.AddGroup()
			common.Info.Save()
		},
	}
}

// reads information about the Group
// in fact, it reads the header of the .kps file correspondent to this Group
func DescGroup() *cobra.Command {
	return &cobra.Command{
		Use:   "desc [name]",
		Short: "describes the group",
		Args:  cobra.ExactArgs(1),
		Run: func(Cmd *cobra.Command, args []string) {
			f, err := GetFileForGroup(args[0], os.O_RDONLY, 0600, false)
			if err != nil {
				common.ShowError(err.Error(), 7)
			}
			defer f.Close()

			var g Group

			err = binary.Read(f, binary.BigEndian, &g)

			if err != nil {
				panic("can't read file, it may be corrupted")
			}

			g.show()
		},
	}
}

func CreateNoteInGroup() *cobra.Command {
	return &cobra.Command{
		Use:   "add [group] [note]",
		Short: "add a note to a group",
		Args:  cobra.ExactArgs(2),
		Run: func(Cmd *cobra.Command, args []string) {
			file, err := GetFileForGroup(args[0], os.O_RDWR, 0666, false)
			if err != nil {
				common.ShowError(err.Error(), 7)
			}
			defer file.Close()

			var g Group

			if err := binary.Read(file, binary.BigEndian, &g); err != nil {
				common.ShowError(err.Error(), 12)
			}

			// how many to skip: size(Group) + size(note) * (how many notes already saved)
			// this is due the fact that any .kps file starts with an header holding
			// information about the file and Group in which it contains. So, we have
			// to skip this header (type Group) and begin to count from it
			skip := int64(binary.Size(&g)) + (int64(binary.Size(&notes.Note{})) * int64(g.Size))

			if _, err := file.Seek(skip, 0); err != nil {
				common.ShowError(err.Error(), 12)
			}

			n := common.GenerateNoteForGroup(args[1], int64(g.Size))

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

func GetFileForGroup(groupName string, flag int, mode fs.FileMode, skip bool) (*os.File, error) {
	fileName := groupName + ".kps"
	fileName, err := common.FixPath(fileName)

	if err != nil {
		return nil, fmt.Errorf("invalid name for file")
	}

	if !common.DoesFileExists(fileName) {
		return nil, fmt.Errorf("there is no Group with given name")
	}

	f, err := os.OpenFile(fileName, flag, mode)
	if skip {
		if _, err := f.Seek(int64(binary.Size(&Group{})), 0); err != nil {
			common.ShowError(err.Error(), 12)
		}
	}
	return f, err
}

func ReadNotesInGroup() *cobra.Command {
	return &cobra.Command{
		Use:     "all [name]",
		Short:   "reads all notes from group with [name]",
		Aliases: []string{"view"},
		Args:    cobra.ExactArgs(1),
		Run: func(Cmd *cobra.Command, args []string) {
			file, err := GetFileForGroup(args[0], os.O_RDONLY, 0600, true)
			if err != nil {
				common.ShowError(err.Error(), 7)
			}
			defer file.Close()

			var n notes.Note

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
					n.Show()
				}
			}
		},
	}
}

func ReadNoteFromGroup() *cobra.Command {
	return &cobra.Command{
		Use:     "read [group] [noteId]",
		Short:   "read single note from group by its id",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{},
		Run: func(Cmd *cobra.Command, args []string) {
			file, err := GetFileForGroup(args[0], os.O_RDONLY, 0600, true)
			if err != nil {
				common.ShowError(err.Error(), 7)
			}
			defer file.Close()

			id, err := strconv.ParseInt(args[1], 10, 64)

			if err != nil {
				common.ShowError("invalid id!", 14)
			}

			var n notes.Note

			for {
				err := binary.Read(file, binary.BigEndian, &n)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					} else {
						panic("keep: error while reading binary file")
					}
				}
				if n.Id == id {
					n.Show()
					break
				}
			}
		},
	}
}

func GetGroups() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "show all groups",
		Run: func(Cmd *cobra.Command, args []string) {
			entries, err := os.ReadDir(common.NotesFolder) // reading current folder
			if err != nil {
				common.ShowError(fmt.Sprintf("can't read notes folder! - err: %v", err.Error()), 21)
			}

			for _, e := range entries {
				if !e.IsDir() {
					info, err := e.Info()
					if err != nil {
						panic(err)
					}

					name := strings.Split(info.Name(), ".")[0] // the name of the file without any extension

					if path.Ext(info.Name()) == ".kps" && name != "keeps" { // "keeps" is the name of the default file which doesn't is in any Group
						g, err := ReadFileHeader(strings.Split(info.Name(), ".")[0])
						if err != nil && !errors.Is(err, io.EOF) {
							panic("keep: there was an error while reading header! - error: " + err.Error())
						}
						g.Short()
					}
				}
			}
		},
	}
}

func ReadFileHeader(fileName string) (Group, error) {
	f, err := GetFileForGroup(fileName, os.O_RDONLY, 0600, false)
	if err != nil {
		return Group{}, err
	}
	defer f.Close()

	var g Group

	err = binary.Read(f, binary.BigEndian, &g)

	if err != nil {
		return Group{}, err
	}

	return g, nil
}
