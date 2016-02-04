package alldebrid

import (
	"bufio"
	"github.com/howeyc/gopass"
	"io/ioutil"
	"os"
	"path/filepath"
)

func getCredentials() (string, string) {
	print("Login: ")
	bio := bufio.NewReader(os.Stdin)
	id, _, _ := bio.ReadLine()

	print("Password: ")
	pass, _ := gopass.GetPasswdMasked()

	return string(id), string(pass)
}

func getFileContent(fullname string) (string, error) {
	if contents, err := ioutil.ReadFile(fullname); err != nil {
		return "", err
	} else {
		return string(contents), nil
	}
}

func getFullName(filename string) string {
	fp := string(filepath.Separator)
	path := os.Getenv("HOME") + fp + ".config" + fp + "alldebrid"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0700)
	}

	return path + fp + filename
}

func btos(value bool) string {
	if value {
		return "1"
	} else {
		return "0"
	}
}
