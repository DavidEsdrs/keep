package configs_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/DavidEsdrs/keep/common"
	"github.com/DavidEsdrs/keep/configs"
	"github.com/DavidEsdrs/keep/utils"
)

func TestCreateInfoFile(t *testing.T) {
	filePath, err := utils.GetKeepFilePath()
	if err != nil {
		t.Fatal(err)
	}
	targetFile := path.Join(filePath, common.INFO_FILE_PATH)

	t.Run("Create info file", func(t *testing.T) {
		defer os.Remove(targetFile)
		_, err := configs.Create(targetFile)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("test full creation", func(t *testing.T) {
		defer os.Remove(targetFile)
		_, err := configs.CreateInfoFile(common.INFO_FILE_PATH)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Parse info content", func(t *testing.T) {
		defer os.Remove(targetFile)
		notesInfoWritten, err := configs.Create(targetFile)
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(targetFile)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		notesInfoRetrieved, err := configs.ParseInfoContent(f)
		if err != nil {
			t.Fatal(err)
		}
		if *notesInfoWritten != notesInfoRetrieved {
			t.Fatal("data written is different of retrieved data")
		}
	})
}

func TestNoteInfoFileSave(t *testing.T) {
	noteInfo := createDefaultNoteInfo(time.Now())
	err := noteInfo.Save()
	if err != nil {
		t.Fatal(err)
	}
}

func createDefaultNoteInfo(t time.Time) configs.NotesInfo {
	var (
		title       [20]rune
		description [200]rune
	)

	copy(title[:], []rune("default"))
	copy(description[:], []rune("this is the default note group - notes with no group given will be stored here"))

	now := t.UnixMilli()

	notesInfo := configs.NotesInfo{
		Title:       title,
		Description: description,
		Size:        0,
		SizeAlltime: 0,
		LastUpdate:  now,
		CreatedAt:   now,
	}
	return notesInfo
}
