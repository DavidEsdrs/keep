package groups

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"

	"github.com/DavidEsdrs/keep/notes"
	"github.com/DavidEsdrs/keep/utils"
)

func CreateGroup(groupName, description string) {
	_, err := NewNoteFile(groupName, description)
	if err != nil {
		panic(err)
	}
	fmt.Printf("group %s created\n", groupName)
}

func NewNoteFile(title, description string) (notes.NoteFileHeader, error) {
	var nfh notes.NoteFileHeader
	kfp, err := utils.GetKeepFilePath()
	if err != nil {
		return nfh, err
	}
	noteFilepath := path.Join(kfp, title+".kps")
	f, err := os.OpenFile(noteFilepath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nfh, err
	}
	defer f.Close()
	header := notes.NewNoteFileHeader(title, description, 0, 0)
	err = binary.Write(f, binary.BigEndian, &header)
	return nfh, err
}
