package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"sync"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)


func Secload(c *cli.Context) error {
	CreateOrLoadDir(Root)
	return nil
}

func Secstart(c *cli.Context) error {
	Secload(c)

	if preferences.NeedSetup() {
		color.Red("Warning: not enough services, SecBack will Create 3 Cloud store locally as folders named Cloud")
	var folderStoref FolderStore
	var folderStores FolderStore
	var folderStoret FolderStore
	
	usr, _ := user.Current()
	folderStoref.Path =path.Join(usr.HomeDir, "Cloud_1")
	if !folderStoref.Setup() {
		color.Red("(Local Cloud Store 1) Folder Store: setup incomplete.")
	}
	
	folderStores.Path =path.Join(usr.HomeDir, "Cloud_2")
	if !folderStores.Setup() {
		color.Red("(Local Cloud Store 2) Folder Store: setup incomplete.")
	
	}
	
	folderStoret.Path =path.Join(usr.HomeDir, "Cloud_3")
	if !folderStoret.Setup() {
		color.Red("(Local Cloud Store 3) Folder Store: setup incomplete.")

	}

	preferences.FolderStores = append(preferences.FolderStores, folderStoref)
	preferences.Save()

	color.Green("Success! Added Local Cloud Store 1: %s", folderStoref.Path)
		

	preferences.FolderStores = append(preferences.FolderStores, folderStores)
	preferences.Save()

	color.Green("Success! Added Local Cloud Store 2: %s", folderStores.Path)


	preferences.FolderStores = append(preferences.FolderStores, folderStoret)
	preferences.Save()

	color.Green("Success! Added Local Cloud Store 3: %s", folderStoret.Path)

	}





	
	color.Green("Starting SecBack. Listening on %s", preferences.root)
	StartWatching(preferences.root, preferences.DirMap)

	return nil
}

func Secstatus(c *cli.Context) error {
	Secload(c)

	color.Green("Cloud stores:")
	for i, cs := range preferences.AllCloudStores() {
		fmt.Println(color.GreenString("%v", i+1), cs.Description())
	}
	if preferences.NeedSetup() {
		color.Red("Warning: not enough services.")
	}

	return nil
}

func Secrestore(c *cli.Context) error {
	Secload(c)

	if preferences.NeedSetup() {
		color.Red("Cannot Restore as not enough shares available")
		return nil
	}

	color.Green("Preparing to restore SecBack to %s", preferences.root)
	Restore()

	return nil
}

func Secremove(c *cli.Context) error {
	Secload(c)

	numStores := preferences.RegisteredServices()
	if numStores == 0 {
		color.Red("There are no cloud stores to delete.")
		return nil
	}

	color.Green("Cloud stores:")
	for i, cs := range preferences.AllCloudStores() {
		fmt.Println(color.GreenString("%v)", i+1), cs.Description())
	}
	color.Cyan("Enter the number of the store you would like to remove:")

	var d int
	for true {
		_, err := fmt.Scanf("%d", &d)
		if err != nil || d < 0 || d > numStores {
			color.Red("Please enter a number between %v and %v", 1, numStores)
		} else {
			break
		}
	}

	if d <= len(preferences.FolderStores) {
		ind := d - 1
		preferences.FolderStores[ind].Clean()
		preferences.FolderStores = append(preferences.FolderStores[:ind], preferences.FolderStores[ind+1:]...)
		color.Yellow("Deleting Folder Store...")
	} else if d <= len(preferences.FolderStores)+len(preferences.GDriveStores) {
		ind := d - 1 - len(preferences.FolderStores)
		preferences.GDriveStores[ind].Clean()
		preferences.GDriveStores = append(preferences.GDriveStores[:ind], preferences.GDriveStores[ind+1:]...)
		color.Yellow("Deleting Google Drive Store...")
	} else {
		ind := d - 1 - len(preferences.FolderStores) - len(preferences.GDriveStores)
		preferences.DropboxStores[ind].Clean()
		preferences.DropboxStores = append(preferences.DropboxStores[:ind], preferences.DropboxStores[ind+1:]...)
		color.Yellow("Deleting Dropbox Store...")
	}

	preferences.Save()
	Secsync(c)

	return nil
}

func Secclean(c *cli.Context) error {
	Secload(c)
	var wg sync.WaitGroup
	for _, cs := range preferences.AllCloudStores() {
		wg.Add(1)
		go func(c CloudStore) {
			defer wg.Done()
			c.Clean()
			color.Green("Done cleaning %v", c.Description())
		}(cs)
	}
	wg.Wait()

	return nil
}

func Secsync(c *cli.Context) error {
	color.Green("Clean:")
	Secclean(c)
	color.Green("Done cleaning.\nBeginning sync:")

	if preferences.NeedSetup() {
		color.Red("Error: not enough services. Cannot sync.")
		return nil
	}

	files, _ := ioutil.ReadDir(preferences.root)
	currentFileMap := make(map[string]bool)
	for _, f := range files {
		if f.Name() == PrefFile {
			continue
		}
		path := path.Join(preferences.root, f.Name())
		currentFileMap[path] = true
		AddFile(path)
	}

	
	for filePath, _ := range preferences.FileMap {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			delete(preferences.FileMap, filePath)
		}
	}

	preferences.Save()
	AddFile(path.Join(preferences.root, PrefFile))

	color.Green("Done syncing.")

	return nil
}



func addFolder(c *cli.Context) error {
	Secload(c)
	var folderStore FolderStore

	if len(c.Args()) < 1 {
		color.Red("Error: missing folder path")
		return nil
	}

	folderStore.Path = c.Args()[0]
	if !folderStore.Setup() {
		color.Red("(Cloud Store) Folder Store: setup incomplete.")
		return nil
	}

	preferences.FolderStores = append(preferences.FolderStores, folderStore)
	preferences.Save()

	color.Green("Success! Added folder store: %s", folderStore.Path)
	return nil
}

func addDrive(c *cli.Context) error {
	Secload(c)
	var gdrive GDriveStore

	if (&gdrive).Setup() == false {
		color.Red("(Cloud Store) Google Drive: setup incomplete.")
		return nil
	}

	
	preferences.GDriveStores = append(preferences.GDriveStores, gdrive)
	preferences.Save()

	color.Green("Success! Added Google Drive Store.")

	return nil
}

func addDropbox(c *cli.Context) error {
	Secload(c)

	var dropbox DropboxStore

	if (&dropbox).Setup() == false {
		color.Red("(Cloud Store) Dropbox: setup incomplete.")
		return nil
	}



	preferences.DropboxStores = append(preferences.DropboxStores, dropbox)
	preferences.Save()

	color.Green("Success! Added Dropbox Store.")

	return nil
}


var Root string

func main() {
	app := cli.NewApp()

	app.Name = color.GreenString("SecBack")
	app.Usage = color.GreenString("A Light Weigth Secure Cloud Back")
	app.EnableBashCompletion = true
	app.Version = "Testing"

	usr, _ := user.Current()
	defaultRoot := path.Join(usr.HomeDir, "BackUp")
	Root = defaultRoot

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "root, r",
			Value:       defaultRoot,
			Usage:       "Destination of the secure folder.",
			Destination: &Root,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "start",
			Aliases: nil,
			Usage:   "Start Listening",
			Action:  Secstart,
		},
		{
			Name:    "status",
			Aliases: nil,
			Usage:   "Prints out the current setup.",
			Action:  Secstatus,
		},
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "Add a new cloud store",
			Subcommands: []cli.Command{
				{
					Name:   "folder",
					Usage:  "add folder",
					Action: addFolder,
				},
				{
					Name:   "dropbox",
					Usage:  "add dropbox",
					Action: addDropbox,
				},
				{
					Name:   "drive",
					Usage:  "add google drive",
					Action: addDrive,
				},
			},
		},
		{
			Name:    "restore",
			Aliases: nil,
			Usage:   "Restores",
			Action:  Secrestore,
		},
		{
			Name:    "remove",
			Aliases: nil,
			Usage:   "Removes a cloud store.",
			Action:  Secremove,
		},
		{
			Name:    "clean",
			Aliases: nil,
			Usage:   "Deletes all shares in cloud stores",
			Action:  Secclean,
		},
		{
			Name:    "sync",
			Aliases: nil,
			Usage:   "",
			Action:  Secsync,
		},
	}

	app.Run(os.Args)
}
