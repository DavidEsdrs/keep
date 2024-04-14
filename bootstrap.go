// in this file there are functions that is called in the bootstrapping of the
// CLI
package main

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

func createStoreFile() error {
	if !doesFileExists(filename) {
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}

func createInfoFile() error {
	if doesFileExists(infoFilename) {
		return readFile()
	}

	return createFile()
}

func readFile() error {
	f, err := os.Open(infoFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	// ler as informações binárias
	var notesInfo NotesInfo
	err = binary.Read(f, binary.BigEndian, &notesInfo)

	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	// assinalar ao objeto global
	info = &notesInfo

	return nil
}

func createFile() error {
	// criar o arquivo
	f, err := os.Create(infoFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	// escrever informações primárias
	notesInfo := NotesInfo{}

	err = binary.Write(f, binary.BigEndian, &notesInfo)

	// assinalar ao objeto global
	info = &notesInfo

	return err
}
