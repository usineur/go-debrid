package alldebrid

import (
	"crypto/tls"
	"fmt"
	"github.com/andelf/go-curl"
	"github.com/usineur/goch"
	"io"
	"os"
	"strings"
)

const host = "https://www.alldebrid.com"

var cookie string = getFullName("cookie.txt")

type passThru struct {
	io.Reader
	total float64
}

func sendRequest(path string, data map[string]string, form interface{}) (string, string, error) {
	easy := curl.EasyInit()
	defer easy.Cleanup()

	if url, err := goch.EncodeUrl(host, path, data); err != nil {
		return "", "", err
	} else {
		doc := ""

		if form != nil {
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
}

func (pt *passThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.total += float64(n)

	if err == nil {
		fmt.Printf("Read %v bytes for a total of %.2fMB \r", n, pt.total/1048576)
	}

	return n, err
}

func netcat(dst io.Writer, url string) error {
	config := &tls.Config{}

	if host, path, err := goch.DecodeUrl(url); err != nil {
		return err
	} else if conn, err := tls.Dial("tcp", host+":443", config); err != nil {
		return err
	} else {
		str := fmt.Sprintf("GET %v HTTP/1.0\r\nHost: %v\r\n\r\n", path, host)
		go io.Copy(conn, strings.NewReader(str))
		_, err := io.Copy(dst, &passThru{Reader: conn})

		return err
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
