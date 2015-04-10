package alldebrid

import (
	"fmt"
	"github.com/andelf/go-curl"
	"github.com/usineur/goch"
	"os"
	"strings"
)

const host = "https://www.alldebrid.com"

var cookie string = getFullName("cookie.txt")

func sendRequest(path string, data map[string]string, form interface{}) (string, string, error) {
	easy := curl.EasyInit()
	defer easy.Cleanup()

	doc := ""
	url := ""

	if url = host + path + "?" + goch.PrepareFields(data); form != nil {
		url = strings.Replace(url, "www", "upload", -1)
		easy.Setopt(curl.OPT_HTTPPOST, form)
	}

	easy.Setopt(curl.OPT_URL, url)
	easy.Setopt(curl.OPT_COOKIEFILE, cookie)
	if path == "/register/" {
		easy.Setopt(curl.OPT_COOKIEJAR, cookie)
	}
	easy.Setopt(curl.OPT_VERBOSE, false)
	easy.Setopt(curl.OPT_FOLLOWLOCATION, true)
	easy.Setopt(curl.OPT_WRITEFUNCTION, func(content []byte, _ interface{}) bool {
		doc += string(content)
		return true
	})

	if err := easy.Perform(); err != nil {
		return "", "", err
	} else if eff, err := easy.Getinfo(curl.INFO_EFFECTIVE_URL); err != nil {
		return "", "", err
	} else {
		return doc, eff.(string), nil
	}
}

func getCookie() error {
	id, pass := getCredentials()
	fields := map[string]string{
		"action":         "login",
		"login_login":    id,
		"login_password": pass,
	}

	if res, eff, err := sendRequest("/register/", fields, nil); err != nil {
		return err
	} else if eff != host+"/" {
		os.Remove(cookie)

		if form, err := goch.GetFormValues(res, "//form[@name=\"connectform\"]"); err != nil {
			return err
		} else if captcha, exist := form["recaptcha_response_field"]; exist && captcha == "manual_challenge" {
			return fmt.Errorf("AllDebrid is asking for a captcha: login to the website first and retry")
		} else {
			return fmt.Errorf("Invalid login/password?")
		}
	} else {
		return nil
	}
}
