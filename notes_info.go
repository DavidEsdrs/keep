package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

type NotesInfo struct {
	NotesQuant uint32
	LastUpdate int64 // UNIX timestamp
	CreatedAt  int64
}

func (n NotesInfo) String() string {
	return fmt.Sprintf("NotesInfo{%v, %v, %v}", info.NotesQuant, info.LastUpdate, info.CreatedAt)
}

func (n *NotesInfo) Add() {
	n.NotesQuant++
}

func (n *NotesInfo) Remove() {
	n.NotesQuant--
}

func (n *NotesInfo) Save() {
	f, err := os.OpenFile("info.bin", os.O_WRONLY, 0600)
	if err != nil {
		panic("something went wrong with the info file! did you deleted it?")
	}
	defer f.Close()

	if err := binary.Write(f, binary.BigEndian, n); err != nil {
		panic(fmt.Sprintf("%v%v", "something went wrong while writing into the info file!: ", err.Error()))
	}
}

func createFile() error {
	// criar o arquivo
	f, err := os.Create("info.bin")
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

func readFile() error {
	f, err := os.Open("info.bin")
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
