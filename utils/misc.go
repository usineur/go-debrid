package utils

import (
	"bufio"
	"github.com/howeyc/gopass"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

func GetCredentials() (string, string) {
	print("Login: ")
	bio := bufio.NewReader(os.Stdin)
	id, _, _ := bio.ReadLine()

	print("Password: ")
	pass, _ := gopass.GetPasswdMasked()

	return string(id), string(pass)
}

func GetFileContent(fullname string) (string, error) {
	if contents, err := ioutil.ReadFile(fullname); err != nil {
		return "", err
	} else {
		return string(contents), nil
	}
}

func GetFullName(filename string) string {
	home := "HOME"
	if IsWindows() {
		home = "USERPROFILE"
	}
	fp := string(filepath.Separator)
	path := os.Getenv(home) + fp + ".config" + fp + "alldebrid"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0700)
	}

	return path + fp + filename
}

func GetChoice(length int) (int, error) {
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

func Btos(value bool) string {
	return strconv.FormatBool(value)
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}
