package main

import (
	"crypto/sha256"
	"encoding/base64"
)


const (
	GoogleDriveClientSecret = string(`{"installed":{"client_id":"101907208653-aet8icjmf9ijkhu5398hf6srm2adni26.apps.googleusercontent.com","project_id":"tough-mechanic-231115","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"OwYXFaWqQjsAPwL36F1ZOlm2","redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}`)
	DropboxClientKey        = "6cjd40wz0bim84q"
	DropboxClientSecret     = "vy1i9ulpgucrigp"
)



func check(e error) {
	if e != nil {
		panic(e)
	}
}


func SHA256Base64URL(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func checkSHA2(hash string, data []byte) bool {
	return SHA256Base64URL(data) == hash
}
