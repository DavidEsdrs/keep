package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

type note struct {
	Id        int64
	Text      [100]byte
	Color     int8
	CreatedAt int64
}

func (n note) show() {
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
