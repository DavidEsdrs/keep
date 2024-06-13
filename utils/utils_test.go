package utils_test

import (
	"bufio"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/DavidEsdrs/keep/utils"
)

func TestGetKeepFilePath(t *testing.T) {
	_, err := utils.GetKeepFilePath()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoesFileExists(t *testing.T) {
	filename := "this_file_exists"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if !utils.DoesFileExists(filename) {
		t.Fatal()
	} else {
		os.Remove(filename)
	}
}

func TestOpenOrCreate(t *testing.T) {
	filename := "TestOpenOrCreate"

	t.Run("Read only file", func(t *testing.T) {
		defer os.Remove(filename)
		f, err := utils.OpenOrCreate(filename, os.O_CREATE|os.O_RDONLY, 0400)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		w.WriteString("this is illegal write")
		err = w.Flush() // flushing ensures we are trying to write into the file
		if err == nil {
			t.Fatal("illegal write to read-only file")
		}

		r := bufio.NewReader(f)
		if _, err := r.ReadByte(); err != nil && !errors.Is(err, io.EOF) {
			t.Fatal("unable to read file - ", err)
		}
	})

	t.Run("Open already existing file", func(t *testing.T) {
		defer os.Remove(filename)
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()

		f, err = utils.OpenOrCreate(filename, os.O_RDONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
	})
}
