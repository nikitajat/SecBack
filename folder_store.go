package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/fatih/color"
)

type FolderStore struct {
	Path string `json:"path"`
}


func (f FolderStore) Setup() bool {
	for _, fs := range preferences.FolderStores {
		if fs.Path == f.Path {
			color.Red("Folder store at %v already exists.", f.Path)
			return false
		}
	}
	os.MkdirAll(f.Path, 0777)
	return true
}


func (f FolderStore) Upload(share Share) {
	sharePath := path.Join(f.Path, string(share.SID))
	err := ioutil.WriteFile(sharePath, share.Data, 0770)
	if err != nil {
		color.Red("Error: %s", err)
		return
	}

	color.Magenta("Share %s saved successfully!", sharePath)
}


func (f FolderStore) Delete(sid ShareID) {
	sharePath := f.Path + "/" + string(sid)
	if _, err := os.Stat(sharePath); err != nil {
		color.Red("Share %s does not exist.", sharePath)
		return
	}

	err := os.Remove(sharePath)
	if err != nil {
		color.Red("Error: could not delete file. %s", err)
		return
	}

	color.Yellow("Share %s deleted successfully!", sid)
}


func (f FolderStore) Restore() string {
	
	return f.Path
}

func (f FolderStore) Description() string {
	label := "Folder store at " + f.Path

	files, _ := ioutil.ReadDir(f.Path)
	for _, f := range files {
		label += fmt.Sprintf("\n\t%s %s", color.YellowString("-"), f.Name())
	}

	return label
}

func (f FolderStore) ShortDescription() string {
	return "Folder store: " + f.Path
}


func (f FolderStore) Clean() {
	files, _ := ioutil.ReadDir(f.Path)
	for _, file := range files {
		color.Yellow("Removing Folder Store: %v", file.Name())
		os.Remove(path.Join(f.Path, file.Name()))
	}
}
