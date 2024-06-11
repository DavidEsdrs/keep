package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path"
	"testing"
)

func TestGetKeepFilePath(t *testing.T) {
	_, err := GetKeepFilePath()
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

	if !doesFileExists(filename) {
		t.Fatal()
	} else {
		os.Remove(filename)
	}
}

func TestCreateInfoFile(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	filename := "info.ksp"
	dir := path.Join(homeDir, ".keep", filename)

	t.Run("Info file doesn't exist", func(t *testing.T) {
		if doesFileExists(dir) {
			err = os.Remove(dir)
			if err != nil {
				t.Fatal(err)
			}
		}
		// main test
		_, err = createInfoFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		if !doesFileExists(dir) {
			t.Fatal("info file not created!")
		}
	})

	t.Run("File does exist", func(t *testing.T) {
		if !doesFileExists(dir) {
			// we hard code file to test the case the file exists
			f, err := os.OpenFile(dir, os.O_CREATE|os.O_RDWR, 0600)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			w := bufio.NewWriter(f)
			_, err = w.WriteString("0,0,0")
			if err != nil {
				t.Fatal(err)
			}
		}
		_, err = createInfoFile(filename)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestOpenOrCreate(t *testing.T) {
	filename := "TestOpenOrCreate"

	t.Run("Read only file", func(t *testing.T) {
		defer os.Remove(filename)
		f, err := OpenOrCreate(filename, os.O_CREATE|os.O_RDONLY, 0400)
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

		f, err = OpenOrCreate(filename, os.O_RDONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
	})
}
