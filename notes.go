package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/fatih/color"
)

type NoteFileHeader struct {
	Title       [20]rune
	Description [200]rune
	Size        uint32
	SizeAlltime uint32
	CreatedAt   int64 // timestamp
}

func (n *NoteFileHeader) Show() {
	c := color.New(color.BgHiWhite).Add(color.Bold)

	t := time.Unix(n.CreatedAt/1000, 0)

	blue := color.New(color.BgHiBlue).Add(color.Bold)

	blue.DisableColor()

	blue.Print(string(n.Title[:]) + " ~ ")

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

	c.Printf(" %s ", string(n.Description[:]))
	c.Println()
}

func NewNoteFileHeader(t, d string, size, sizeAllTime uint32) NoteFileHeader {
	var (
		title       [20]rune
		description [200]rune
	)

	copy(title[:], []rune(t))
	copy(description[:], []rune(d))

	return NoteFileHeader{
		Title:       title,
		Description: description,
		Size:        size,
		SizeAlltime: sizeAllTime,
		CreatedAt:   time.Now().UnixMilli(),
	}
}

type Note struct {
	Id        int64
	Text      [300]rune
	Color     int32
	CreatedAt int64
}

func NewNote(id int64, text string, c color.Attribute, createAt int64) Note {
	var t [300]rune
	copy(t[:], []rune(text))
	return Note{
		Id:        id,
		Text:      t,
		Color:     int32(c),
		CreatedAt: createAt,
	}
}

func (n Note) Show() {
	c := color.New(color.Attribute(n.Color)).Add(color.Bold)

	t := time.Unix(n.CreatedAt/1000, 0)

	blue := color.New(color.BgHiBlue).Add(color.Bold)

	blue.DisableColor()

	blue.Print(fmt.Sprint(n.Id) + " ~ ")

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
	c.Print(string(n.Text[:]))
	c.Println()
}

// creates a new file named [title].kps with starting values
func NewNoteFile(title, description string) (NoteFileHeader, error) {
	var nfh NoteFileHeader
	kfp, err := GetKeepFilePath()
	if err != nil {
		return nfh, err
	}
	noteFilepath := path.Join(kfp, title+".kps")
	f, err := os.OpenFile(noteFilepath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nfh, err
	}
	defer f.Close()
	header := NewNoteFileHeader(title, description, 0, 0)
	err = binary.Write(f, binary.BigEndian, &header)
	return nfh, err
}

func AddNote(groupname string, text string) error {
	var nfh NoteFileHeader

	kfp, err := GetKeepFilePath()
	if err != nil {
		return err
	}

	noteFilepath := path.Join(kfp, groupname+".kps")
	f, err := os.OpenFile(noteFilepath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("\"%s\" group not found", groupname)
	}
	defer f.Close()

	err = binary.Read(f, binary.BigEndian, &nfh)
	if err != nil {
		return err
	}

	_, err = f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	note := NewNote(int64(nfh.SizeAlltime)+1, text, randomColor(), time.Now().UnixMilli())

	err = binary.Write(f, binary.BigEndian, &note)
	if err != nil {
		return err
	}

	nfh.Size++
	nfh.SizeAlltime++

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = binary.Write(f, binary.BigEndian, &nfh)
	return err
}

func GetGroupHeader(groupName string) (NoteFileHeader, error) {
	var nfh NoteFileHeader
	kfp, err := GetKeepFilePath()
	if err != nil {
		return nfh, err
	}
	noteFilepath := path.Join(kfp, groupName+".kps")
	f, err := os.OpenFile(noteFilepath, os.O_RDWR, 0)
	if err != nil {
		return nfh, err
	}
	defer f.Close()
	err = binary.Read(f, binary.BigEndian, &nfh)
	return nfh, err
}

// ReadAllNotes emits all notes stored within a .kps file
func ReadAllNotes(filename string) (<-chan Note, error) {
	var (
		nfh NoteFileHeader
		out = make(chan Note)
	)
	kfp, err := GetKeepFilePath()
	if err != nil {
		return nil, err
	}
	noteFilepath := path.Join(kfp, filename)
	f, err := os.OpenFile(noteFilepath, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("no group with given name")
	}
	_, err = f.Seek(int64(binary.Size(nfh)), io.SeekStart)
	if err != nil {
		return nil, err
	}
	go func() {
		var n Note
		for {
			err = binary.Read(f, binary.BigEndian, &n)
			if err != nil {
				break
			}
			if n.Id > 0 {
				out <- n
			}
		}
		f.Close()
		close(out)
	}()
	return out, err
}

func GetNoteById(groupName string, id int64) (Note, error) {
	var (
		result Note
		nfh    NoteFileHeader
	)
	kfp, err := GetKeepFilePath()
	if err != nil {
		return result, err
	}

	noteFilepath := path.Join(kfp, groupName+".kps")
	f, err := os.OpenFile(noteFilepath, os.O_RDONLY, 0)
	if err != nil {
		return result, err
	}
	defer f.Close()

	_, err = f.Seek(int64(binary.Size(nfh))+int64(binary.Size(result))*(id-1), io.SeekStart)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return result, fmt.Errorf("invalid id")
		}
		return result, fmt.Errorf("invalid id %v for range: %w", id, err)
	}

	err = binary.Read(f, binary.BigEndian, &result)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return result, fmt.Errorf("invalid id")
		}
		return result, err
	}

	if result.Id != id {
		return result, fmt.Errorf("unexpected entity got from given id")
	}

	return result, nil
}

func DeleteNoteById(groupName string, id int64) error {
	var (
		nfh NoteFileHeader
	)
	kfp, err := GetKeepFilePath()
	if err != nil {
		return err
	}

	noteFilepath := path.Join(kfp, groupName+".kps")
	f, err := os.OpenFile(noteFilepath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Seek(int64(binary.Size(nfh))+int64(binary.Size(Note{}))*(id-1), io.SeekStart); err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("invalid id")
		}
		return fmt.Errorf("invalid id %v for range: %w", id, err)
	}

	nfh.Size--

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if err := binary.Write(f, binary.BigEndian, &nfh); err != nil {
		return err
	}

	return binary.Write(f, binary.BigEndian, &Note{Id: -1})
}

func DeleteGroup(groupName string) error {
	kfp, err := GetKeepFilePath()
	if err != nil {
		return err
	}

	noteFilepath := path.Join(kfp, groupName+".kps")

	if !doesFileExists(noteFilepath) {
		return fmt.Errorf("no group with given name")
	}

	return os.Remove(noteFilepath)
}

func GetGroups() ([]NoteFileHeader, error) {
	var groups []NoteFileHeader

	keepFilePath, err := GetKeepFilePath()
	if err != nil {
		return groups, err
	}

	entries, err := os.ReadDir(keepFilePath)
	if err != nil {
		return groups, fmt.Errorf("unable to read dir: %w", err)
	}

	for _, e := range entries {
		if isKpsFile(e) {
			header, err := GetKpsHeader(path.Join(keepFilePath, e.Name()))
			if err == nil {
				groups = append(groups, header)
			}
		}
	}

	return groups, nil
}

func isKpsFile(entry fs.DirEntry) bool {
	return !entry.IsDir() && ExtractExtension(entry.Name()) == "kps"
}

// GetKpsHeader returns the header of a .kps binary file and a nil error if it has success.
func GetKpsHeader(filename string) (NoteFileHeader, error) {
	var header NoteFileHeader
	f, err := os.Open(filename)
	if err != nil {
		return header, err
	}
	if err := binary.Read(f, binary.BigEndian, &header); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return header, nil
		}
		return header, fmt.Errorf("unable to read file header: %w", err)
	}
	return header, nil
}

func CreateSingleNote(text string) error {
	var nfh NoteFileHeader
	kfp, err := GetKeepFilePath()
	if err != nil {
		return err
	}

	noteFilepath := path.Join(kfp, filename+".kps")
	f, err := os.OpenFile(noteFilepath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := binary.Read(f, binary.BigEndian, &nfh); err != nil {
		return err
	}

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	note := NewNote(int64(nfh.SizeAlltime)+1, text, randomColor(), time.Now().UnixMilli())

	if err := binary.Write(f, binary.BigEndian, &note); err != nil {
		return err
	}

	nfh.Size++
	nfh.SizeAlltime++

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if err := binary.Write(f, binary.BigEndian, &nfh); err != nil {
		return err
	}

	info.Add()
	info.Save()

	return nil
}
