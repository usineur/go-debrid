package alldebrid

import (
	"bufio"
	"github.com/howeyc/gopass"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
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

func getChoice(length int) (int, error) {
	print("? ")
	bio := bufio.NewReader(os.Stdin)
	num, _, _ := bio.ReadLine()

	if res, err := strconv.Atoi(string(num)); err != nil {
		return -1, err
	} else if res < 0 || res > length-1 {
		return -1, nil
	} else {
		return res, nil
	}
}

func btos(value bool) string {
	return strconv.FormatBool(value)
}
