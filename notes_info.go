package main

import (
	"encoding/binary"
	"fmt"
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
	f, err := os.OpenFile(infoFilename, os.O_WRONLY, 0600)
	if err != nil {
		panic("something went wrong with the info file! did you deleted it?")
	}
	defer f.Close()

	if err := binary.Write(f, binary.BigEndian, n); err != nil {
		panic(fmt.Sprintf("%v%v", "something went wrong while writing into the info file!: ", err.Error()))
	}
}
