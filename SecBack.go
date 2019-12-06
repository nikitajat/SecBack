package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/fatih/color"
)






type CloudStore interface {
	Upload(share Share)
	Delete(sid ShareID)


	Restore() string

	Description() string
	
	Clean()
}

type FileShare struct {
	SID  ShareID `json:"sid"`
	Hash string  `json:"hash"` 
}


type Pref struct {
	root string


	FolderStores []FolderStore `json:"folder_stores"`

	GDriveStores []GDriveStore `json:"gdrive_stores"`

	DropboxStores []DropboxStore `json:"dropbox_stores"`

	
	FileMap map[string]FileShare `json:"files"`

	
	DirMap map[string]bool `json:"dirs"`
}


func (p Pref) RegisteredServices() int {
	return len(p.FolderStores) + len(p.GDriveStores) + len(p.DropboxStores)
}


func (p Pref) NeedSetup() bool {
	return p.RegisteredServices() < 2
}



func (p Pref) AllCloudStores() []CloudStore {

	cloudStores := make([]CloudStore, p.RegisteredServices())

	ind := 0
	for _, fs := range p.FolderStores {
		cloudStores[ind] = CloudStore(fs)
		ind += 1
	}

	for _, gds := range p.GDriveStores {
		cloudStores[ind] = CloudStore(gds)
		ind += 1
	}

	for _, dbs := range p.DropboxStores {
		cloudStores[ind] = CloudStore(dbs)
		ind += 1
	}

	return cloudStores
}


func (p Pref) Save() {
	FilePath := path.Join(p.root, PrefFile)
	FileBytes, err := json.MarshalIndent(preferences, "", "    ")
	check(err)

	ioutil.WriteFile(FilePath, FileBytes, 0660)
}



var preferences Pref

const PrefFile = ".config"
const IgnoreFile = ".ignore"



func CreateOrLoadDir(root string) {
	os.MkdirAll(root, 0777)

	FilePath := path.Join(root, PrefFile)
	FileBytes, err := ioutil.ReadFile(FilePath)
	if err != nil {
		color.Green("Creating new .config in secure folder")
		preferences.DirMap = make(map[string]bool)
		preferences.FileMap = make(map[string]FileShare)
		preferences.FileMap[FilePath] = FileShare{SID: ShareID(PrefFile), Hash: ""}
	} else {
		json.Unmarshal(FileBytes, &preferences)
	}

	IgnorePath := path.Join(root, IgnoreFile)
	_, err = ioutil.ReadFile(IgnorePath)
	if err != nil {
		defaultIgnore := []byte(".DS_Store\n")

		preferences.FileMap[IgnorePath] = FileShare{SID: ShareID(IgnoreFile), Hash: SHA256Base64URL(defaultIgnore)}

		

		errWrite := ioutil.WriteFile(IgnorePath, defaultIgnore, 0777)
		if errWrite != nil {
			color.Red("Error: could not write to %s: %s", FilePath, errWrite)
		}
	}

	preferences.root = root
	preferences.Save()
}



func IsValidPath(filePath string) bool {
	base := filepath.Base(filePath)
	IgnorePath := path.Join(preferences.root, IgnoreFile)
	Ignore, err := os.Open(IgnorePath)
	if err != nil {
		return true
	}

	scanner := bufio.NewScanner(Ignore)
	for scanner.Scan() {
		pattern := scanner.Text()
		
		ok, err := filepath.Match(pattern, base)

		if ok {
			return false
		}
		if err != nil {
			fmt.Println(err)
		}
	}

	
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	return true
}


func AddFile(filePath string) {
	if !IsValidPath(filePath) {
		color.Blue("Path %s is in .ignore. No actions will be performed.", filePath)
		return
	}
	file, _ := os.Open(filePath)
	fi, err := file.Stat()
	if err != nil {
		color.Red("Cannot get file info: %s", err)
		return
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		files, _ := ioutil.ReadDir(filePath)
		preferences.DirMap[path.Clean(filePath)] = true

		for _, f := range files {
			AddFile(path.Join(filePath, f.Name()))
		}
		return
	case mode.IsRegular():
		break
	}

	var sid ShareID
	if existingFileShare, ok := preferences.FileMap[filePath]; ok {
		sid = existingFileShare.SID
	} else {
		
		sid = RandomShareID()
	}

	
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		color.Red("Cannot read file: %s", err)
		return
	}

	preferences.FileMap[filePath] = FileShare{SID: sid, Hash: SHA256Base64URL(fileBytes)}


	allCloudStores := preferences.AllCloudStores()
	shares := CreateShares(fileBytes, sid, len(allCloudStores))

	
	for i, cs := range allCloudStores {
		cs.Upload(shares[i])
	}

	
	if sid != ShareID(".config") {
		preferences.Save()
	}

}

func DeleteFile(filePath string) {
	if !IsValidPath(filePath) {
		color.Red("Path %s is in .ignore. No actions will be performed.", filePath)
		return
	}

	potenDirPath := path.Clean(filePath)
	if _, ok := preferences.DirMap[potenDirPath]; ok {
		DeleteDir(potenDirPath)
		return
	}

	allCloudStores := preferences.AllCloudStores()

	if fileShare, ok := preferences.FileMap[filePath]; ok {
	
		for _, cs := range allCloudStores {
			cs.Delete(fileShare.SID)
		}

		delete(preferences.FileMap, filePath)
		preferences.Save()

		color.Yellow("Deleted share from all cloud stores.")
		return
	}

	color.Red("Path %s is not tracked. Cannot find share id.", filePath)
}

func DeleteDir(dirPath string) {


	delete(preferences.DirMap, dirPath)

	for filePath, _ := range preferences.FileMap {
		dirMatch, _ := path.Split(filePath)
		if path.Clean(dirMatch) != path.Clean(dirPath) {
			continue
		}
		DeleteFile(filePath)
	}
}


func Restore() {
	allCloudStores := preferences.AllCloudStores()
	sharePaths := make([]string, len(allCloudStores))

	
	for i, cs := range allCloudStores {
		sp := cs.Restore()
		if sp == "" {
			color.Red("Restore failed for %v", cs)
			return
		}
		sharePaths[i] = sp
	}

	FileBytes := restoreShareID(ShareID(PrefFile), sharePaths)
	var restoredPrefs Pref
	err := json.Unmarshal(FileBytes, &restoredPrefs)
	if err != nil {
		color.Red("Cannot restore file from cloud services.")
		return
	}

	
	for dirPath, _ := range restoredPrefs.DirMap {
		os.MkdirAll(dirPath, 0777)
		preferences.DirMap[dirPath] = true
	}

	
	for filePath, fileShare := range restoredPrefs.FileMap {
		fileBytes := restoreShareID(fileShare.SID, sharePaths)
		if len(fileBytes) == 0 {
			continue
		}

		if fileShare.SID != ShareID(PrefFile) && checkSHA2(fileShare.Hash, fileBytes) == false {
			color.Red("Error: invalid SHA2 checksum for share %s. Skipping.", fileShare.SID)
			continue
		}

		err := ioutil.WriteFile(filePath, fileBytes, 0777)
		if err != nil {
			color.Red("Error writing restored file %s: %s", filePath, err)
		}
	}
	color.Green("Done. Restored all files!")
}

func restoreShareID(sid ShareID, sharePaths []string) []byte {
	fileShares := make([]Share, len(sharePaths))

	sharesFound := 0
	for i, sp := range sharePaths {
			
		file := path.Join(sp, string(sid))
		dataBytes, err := ioutil.ReadFile(file)
		if err != nil {
			color.Red("(Skipping share) Cannot read file %s: %s", file, err)
			continue
		}

		fileShares[i] = Share{SID: sid, Data: dataBytes}
		sharesFound++
	}

	if sharesFound < preferences.RegisteredServices() {
		color.Red("Couldn't retrieve enough shares to restore %s", sid)
		return []byte{}
	} else {
		return CombineShares(fileShares)
	}
}
