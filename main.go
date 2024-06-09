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

	"github.com/DavidEsdrs/godeline"
	editnode "github.com/DavidEsdrs/godeline/edit-node"
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

	t := editnode.NewEditTree()

	t.AddDelimiterType("{", "}")
	t.AddDelimiterType("[", "]")
	t.AddDelimiterType("(", ")")
	t.AddDelimiterType("`", "`")

	p := godeline.NewProcessor(&t, 1<<12)

	p.Sanitize()

	processor = &p
}

func main() {
	// create is the base command, i.e, when the CLI is called with no
	// subcommands (such as `keep "this a note"`) it is implicity that we want to
	// create a new note
	rootCmd := create()

	rootCmd.AddCommand(readAll())
	rootCmd.AddCommand(delete())
	rootCmd.AddCommand(readSingle())

	rootCmd.PersistentFlags().Bool("inc", false, "Show the notes in decreasing order")

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

var processor *godeline.Processor

func create() *cobra.Command {
	return &cobra.Command{
		Use:     "[input]",
		Aliases: []string{"create", "add"},
		Short:   "creates a new note",
		Args:    cobra.RangeArgs(1, 10),
		Run: func(cmd *cobra.Command, args []string) {
			dir, err := getKeepFilePath()
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
				} else {
					showError("the length of the note is bigger than allowed!", 10)
				}
			}

			err = w.Flush()
			if err != nil {
				panic("error while flushing! " + err.Error())
			}
			info.Add()
			info.Save()
		},
	}
}

func readAll() *cobra.Command {
	return &cobra.Command{
		Use:     "all",
		Aliases: []string{"remind", "get"},
		Short:   "remind you all notes",
		Run: func(cmd *cobra.Command, args []string) {
			f, err := os.Open(filename)
			if err != nil {
				panic("can't manage to open file! did you deleted it? error: " + err.Error())
			}
			defer f.Close()

			s := bufio.NewScanner(f)

			isIncreasing, _ := cmd.Flags().GetBool("inc")

			if isIncreasing {
				showIncreasingOrder(s)
			} else {
				showDecreasingOrder(s)
			}

			fmt.Printf("%v notes\n", info.NotesQuant)
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
			file, err := OpenOrCreate(filename, os.O_CREATE|os.O_RDWR, 0600)
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
			file, err := OpenOrCreate(filename, os.O_CREATE|os.O_RDONLY, 0600)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			id, _ := strconv.ParseInt(args[0], 10, 64)

			s := bufio.NewScanner(file)

			for s.Scan() {
				line := s.Text()

				note := parseTextAsNote(line)

				if note.id == id {
					note.show()
					break
				}
			}
		},
	}
}
