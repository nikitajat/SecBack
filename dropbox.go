package main

import (
	"encoding/json"
	"fmt"	
	//"bytes"
	"os"
	"path"
	"time"
	//"errors"
	"strings"
	"io"
	"io/ioutil"
	"path/filepath"
	"golang.org/x/oauth2"
	"github.com/fatih/color"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/net/context"
)



const (
	configFileName  = "auth.json"
	tokenPersonal   = "personal"
//	tokenTeamAccess = "teamAccess"
//	tokenTeamManage = "teamManage"
)


var (
	personalAppKey      = "6cjd40wz0bim84q"
	personalAppSecret   = "vy1i9ulpgucrigp"
	teamAccessAppKey    = "zud1va492pnehkc"
	teamAccessAppSecret = "p3ginm1gy0kmj54"
	teamManageAppKey    = "xxe04eai4wmlitv"
	teamManageAppSecret = "t8ms714yun7nu5s"
	SS_idc              = ".config"
	SS_id               = ""
)


type DropboxStore struct {
	Dropbox     dropbox.Dropbox `json:"dropbox"`
	AccessToken string
	UserID      int
	SID         string
}

/*func GetClientKeys() (key, secret string) {
	return DropboxClientKey, DropboxClientSecret  //util
}*/
//

var config dropbox.Config

//

type TokenMap map[string]map[string]string

//

func validatePath(p string) (path string, err error) {
	path = p

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	path = strings.TrimSuffix(path, "/")

	return
}

//

func makeRelocationArg(s string, d string) (arg *files.RelocationArg, err error) {
	src, err := validatePath(s)
	if err != nil {
		return
	}
	dst, err := validatePath(d)
	if err != nil {
		return
	}

	arg = files.NewRelocationArg(src, dst)

	return
}

//

func readTokens(filePath string) (TokenMap, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tokens TokenMap
	if json.Unmarshal(b, &tokens) != nil {
		return nil, err
	}

	return tokens, nil
}

//

func writeTokens(filePath string, tokens TokenMap) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Doesn't exist; lets create it
		err = os.MkdirAll(filepath.Dir(filePath), 0700)
		if err != nil {
			return
		}
	}

	// At this point, file must exist. Lets (over)write it.
	b, err := json.Marshal(tokens)
	if err != nil {
		return
	}
	if err = ioutil.WriteFile(filePath, b, 0600); err != nil {
		return
	}
}

//

func oauthConfig(tokenType string, domain string) *oauth2.Config {
	var appKey, appSecret string

		appKey, appSecret = personalAppKey, personalAppSecret

	return &oauth2.Config{
		ClientID:     appKey,
		ClientSecret: appSecret,
		Endpoint:     dropbox.OAuthEndpoint(domain),
	}
}

//

//
const chunkSize int64 = 1 << 24

func uploadChunked(dbx files.Client, r io.Reader, commitInfo *files.CommitInfo, sizeTotal int64) (err error) {
	res, err := dbx.UploadSessionStart(files.NewUploadSessionStartArg(),
		&io.LimitedReader{R: r, N: chunkSize})
	if err != nil {
		fmt.Printf("uploadChunked\n")
		return
	}

	written := chunkSize

	for (sizeTotal - written) > chunkSize {
		cursor := files.NewUploadSessionCursor(res.SessionId, uint64(written))
		args := files.NewUploadSessionAppendArg(cursor)

		err = dbx.UploadSessionAppendV2(args, &io.LimitedReader{R: r, N: chunkSize})
		if err != nil {
			fmt.Printf("uploadChunked2\n")
			return
		}
		written += chunkSize
	}

	cursor := files.NewUploadSessionCursor(res.SessionId, uint64(written))
	args := files.NewUploadSessionFinishArg(cursor, commitInfo)

	if _, err = dbx.UploadSessionFinish(args, r); err != nil {
		fmt.Printf("uploadChunked3\n")		
		return
	}

	return
}

//

func (d *DropboxStore) Setup() bool {
	/*db := dropbox.NewDropbox()
	key, secret := GetClientKeys()
	db.SetAppInfo(key, secret)

	tok, err := getDropboxTokenFromWeb()
	if err != nil {
		color.Red("Unable to get client token: %v", err)
		return false
	}

	err = db.AuthCode(tok)
	if err != nil {
		color.Red("Unable to get client token: %v", err)
		return false
	}

	account, err := db.GetAccountInfo()
	if err != nil {
		color.Red("Unableee to get account information: %v", err)
		return false
	}

	uid := account.UID
	for _, d := range preferences.DropboxStores {
		if d.UserID == uid {
			color.Red("Account for %s already exists.", account.DisplayName)
			return false
		}
	}

	// set the oauth info
	d.Dropbox = *db
	d.AccessToken = db.AccessToken()
	d.UserID = uid
	*/
	domain:=""
		//fmt.Println("initDbx\n")
	dir, err := homedir.Dir()
	if err != nil {
		fmt.Printf("setup 1 error\n")
		return false
	}
	filePath := path.Join(dir, ".config", "dbxcli", configFileName)
	tokType := "personal"
	conf := oauthConfig(tokType, domain)

	tokenMap, err := readTokens(filePath)

	if tokenMap == nil {
		tokenMap = make(TokenMap)
	}
	if tokenMap[domain] == nil {
		tokenMap[domain] = make(map[string]string)
	}
	tokens := tokenMap[domain]

	if err != nil || tokens[tokType] == "" {
		fmt.Printf("1. Go to %v\n", conf.AuthCodeURL("state"))
		fmt.Printf("2. Click \"Allow\" (you might have to log in first).\n")
		fmt.Printf("3. Copy the authorization code.\n")
		fmt.Printf("Enter the authorization code here: ")

		var code string
		if _, err = fmt.Scan(&code); err != nil {
				fmt.Printf("setup 2 error\n")			
				return false
		}
		var token *oauth2.Token
		ctx := context.Background()
		token, err = conf.Exchange(ctx, code)
		if err != nil {
			fmt.Printf("setup 3 error\n")
			return false
		}
		tokens[tokType] = token.AccessToken
		writeTokens(filePath, tokenMap)
	}

	logLevel := dropbox.LogOff

	config = dropbox.Config{tokens[tokType], logLevel, nil, "", domain, nil, nil, nil}

	return true
}

func (d DropboxStore) Upload(share Share) {
	d.SID =string(share.SID)
	if string(share.SID)!=".config"{
		defaultIgnore := []byte(string(share.SID))
		errWrite := ioutil.WriteFile("/home/chinmay/d_u/imp.txt", defaultIgnore, 0777)
		fmt.Println("written")
		if errWrite != nil {
			color.Red("Error: could not write to %s: %s", "/home/chinmay/d_u/imp.txt", errWrite)
				}
			}
	sharePath := path.Join("/home/chinmay/d_u/", string(share.SID))
	err := ioutil.WriteFile(sharePath, share.Data, 0770)
	if err != nil {
		color.Red("making share file Error: %s", err)
		return
	}

	color.Magenta("Share %s saved successfully!", sharePath)

	domain:=""
		//fmt.Println("initDbx\n")
	dir, err := homedir.Dir()
	if err != nil {
		return
	}
	filePath := path.Join(dir, ".config", "dbxcli", configFileName)
	tokType := "personal"
	//conf := oauthConfig(tokType, domain)

	tokenMap, err := readTokens(filePath)

	if tokenMap == nil {
		tokenMap = make(TokenMap)
	}
	if tokenMap[domain] == nil {
		tokenMap[domain] = make(map[string]string)
	}
	tokens := tokenMap[domain]

	//
	logLevel := dropbox.LogOff

	config = dropbox.Config{tokens[tokType], logLevel, nil, "", domain, nil, nil, nil}



	src := sharePath

        dst := "/"+ path.Base(src)
		
	contents, err := os.Open(src)
	if err != nil {
		fmt.Printf("uploading error1\n")
		return
	}
	defer contents.Close()

	contentsInfo, err := contents.Stat()
	if err != nil {

		fmt.Printf("uploading error2\n")
		return
	}

	progressbar := &ioprogress.Reader{
		Reader: contents,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: contentsInfo.Size(),
	}

	commitInfo := files.NewCommitInfo(dst)
	commitInfo.Mode.Tag = "overwrite"

	// The Dropbox API only accepts timestamps in UTC with second precision.
	commitInfo.ClientModified = time.Now().UTC().Round(time.Second)

	dbx := files.New(config)
	if contentsInfo.Size() > chunkSize {
		//return uploadChunked(dbx, progressbar, commitInfo, contentsInfo.Size())
	}

	if _, err = dbx.Upload(commitInfo, progressbar); err != nil {
			fmt.Printf("uploading error 3\n")

		return
	}

	if _, err := os.Stat(sharePath); err != nil {
		color.Red("Share %s does not exist.", sharePath)
		return
	}

	err = os.Remove(sharePath)
	if err != nil {
		color.Red("Error: could not delete file. %s", err)
		return
	}

	//color.Yellow("Share %s deleted successfully!", sid)

	return
}

func (d DropboxStore) Delete(sid ShareID) {
	/*key, secret := GetClientKeys()
	d.Dropbox.SetAppInfo(key, secret)
	d.Dropbox.SetAccessToken(d.AccessToken)

	fmt.Print(color.YellowString("Deleting Dropbox/%s...", sid))

	_, err := d.Dropbox.Delete(string(sid))
	if err != nil {
		color.Red("Error deleting file: ", err)
		return
	}

	//print check mark
	fmt.Print(color.GreenString("\u2713\n"))*/
}

func (d DropboxStore) Description() string {
	return "hello"
}


func (d DropboxStore) Clean() {
return
}

func (d DropboxStore) Restore() string {
	domain:=""
		//fmt.Println("initDbx\n")
	dir, err := homedir.Dir()
	filePath := path.Join(dir, ".config", "dbxcli", configFileName)
	tokType := "personal"
	//conf := oauthConfig(tokType, domain)

	tokenMap, err := readTokens(filePath)

	if tokenMap == nil {
		tokenMap = make(TokenMap)
	}
	if tokenMap[domain] == nil {
		tokenMap[domain] = make(map[string]string)
	}
	tokens := tokenMap[domain]

	//
	logLevel := dropbox.LogOff

	config = dropbox.Config{tokens[tokType], logLevel, nil, "", domain, nil, nil, nil}


	// Default `dst` to the base segment of the source path; use the second argument if provided.
		//dst :="/home/chinmay/d_u/"
	

			//
	restoreDir, err := ioutil.TempDir("", "SecBack_dropbox_restore")
	arg := files.NewDownloadArg(path.Join("/",SS_idc))
		fmt.Println(SS_idc)
	dbx := files.New(config)
	res, contents, err := dbx.Download(arg)
	if err != nil {
		fmt.Println("error restoring .config")
		return restoreDir
	}
	defer contents.Close()
	
	f, err := os.Create(path.Join(restoreDir,"/.config"))
 	if err != nil {
		fmt.Println("error restoring2 ")
		return restoreDir
	}

	progressbar := &ioprogress.Reader{
		Reader: contents,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Downloading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: int64(res.Size),
	}

	if _, err = io.Copy(f, progressbar); err != nil {
			fmt.Println("error last")
			return restoreDir
	}
		//
	buf,err :=ioutil.ReadFile("/home/chinmay/d_u/imp.txt")
		if err!=nil{
			fmt.Println(err)	
		}
	SS_id :=string(buf)
	fmt.Println(SS_id)
	arg1 := files.NewDownloadArg(path.Join("/",SS_id))

	dbx1 := files.New(config)
	res1, contents1, err := dbx1.Download(arg1)
	if err != nil {
		fmt.Println("error restoring share file ")
		return restoreDir
	}
	defer contents1.Close()

	f1, err:= os.Create(path.Join(restoreDir,SS_id))
	//err = os.Chmod(path.Join("/home/chinmay/d_u/",SS_id),0777)
	if err != nil {
		fmt.Println("error restoring 2 share file ")
		return restoreDir
	}


	progressbar1 := &ioprogress.Reader{
		Reader: contents1,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Downloading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: int64(res1.Size),
	}

	if _, err = io.Copy(f1, progressbar1); err != nil {
			fmt.Println("error restoring 2 share file ")
						
	}
			
		return restoreDir
}


