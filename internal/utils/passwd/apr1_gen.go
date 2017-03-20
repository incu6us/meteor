package passwd

import (
	"encoding/base64"
	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/apr1_crypt"
)

type password struct {}

var passwd *password

func (p *password) GenApr1Password(plainPasswd string) string {
	hash, _ := crypt.APR1.New().Generate([]byte(plainPasswd), []byte("$apr1$CvM2O/cg$"))
	return hash
}

func (p *password) GetPasswdForHeader(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

type Password interface {
	GenApr1Password(plainPasswd string) string
	GetPasswdForHeader(username, password string) string
}

func GeneratePassword() Password {
	if passwd == nil {
		passwd = &password{}
	}
	
	return passwd
}