package alldebrid

import (
	"encoding/json"
	"fmt"
	"github.com/andelf/go-curl"
	"github.com/usineur/goch"
	"os"
	"regexp"
	"time"
)

var cookie string = getFullName("cookie.txt")

type service struct {
	Link      string
	Host      interface{} // could be a string or a boolean
	Filename  string
	Icon      string
	Streaming []string
	Nb        int
	Error     string
	Filesize  string
}

func sendRequest(url string, form interface{}, http_code int, jar bool) (string, error) {
	data := ""

	easy := curl.EasyInit()
	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, url)
	easy.Setopt(curl.OPT_COOKIEFILE, cookie)
	easy.Setopt(curl.OPT_VERBOSE, false)
	if form != nil {
		easy.Setopt(curl.OPT_HTTPPOST, form)
	}
	easy.Setopt(curl.OPT_WRITEFUNCTION, func(content []byte, _ interface{}) bool {
		data += string(content)
		return true
	})

	if err := easy.Perform(); err != nil {
		return "", err
	} else if code, err := easy.Getinfo(curl.INFO_HTTP_CODE); err != nil {
		return "", err
	} else if code != http_code {
		return data, fmt.Errorf("Unexpected code: %v, invalid login/password?\n", code)
	} else if jar {
		easy.Setopt(curl.OPT_COOKIEJAR, cookie)
		return data, nil
	} else {
		return data, nil
	}
}

func getCookie() error {
	id, pass := getCredentials()
	url := "https://www.alldebrid.com/register/?action=login&login_login=" + id + "&login_password=" + pass

	if data, err := sendRequest(url, nil, 302, true); err != nil {
		if form, erf := goch.GetFormValues(data, "//form[@name=\"connectform\"]"); erf != nil {
			return erf
		} else if captcha, exist := form["recaptcha_response_field"]; !exist || captcha != "manual_challenge" {
			return err
		} else {
			return fmt.Errorf("AllDebrid is asking for a captcha: login to the website first and retry")
		}
	} else {
		return nil
	}
}

func getUid() (string, error) {
	if contents, err := getFileContent(cookie); err != nil {
		if err := getCookie(); err != nil {
			return "", err
		} else {
			return getUid()
		}
	} else if pattern, err := regexp.Compile(".*uid\t(.*)"); err != nil {
		return "", err
	} else if matches := pattern.FindStringSubmatch(contents); len(matches) != 2 {
		return "", fmt.Errorf("Expected cookie \"uid\" not found\n")
	} else {
		return matches[1], nil
	}
}

func getDownloadLink(link string) (string, string, error) {
	url := "https://www.alldebrid.com/service.php?json=true&link=" + link
	var s service

	if data, err := sendRequest(url, nil, 200, false); err != nil {
		return "", "", err
	} else if data == "login" {
		if err := getCookie(); err != nil {
			return "", "", err
		} else {
			return getDownloadLink(link)
		}
	} else if err := json.Unmarshal([]byte(data), &s); err != nil {
		return "", "", err
	} else if s.Error != "" {
		return "", "", fmt.Errorf(s.Error)
	} else {
		return s.Link, s.Filename, nil
	}
}

func GetTorrentList() error {
	url := "https://www.alldebrid.com/torrent/"

	if data, err := sendRequest(url, nil, 200, false); err != nil {
		if err := getCookie(); err != nil {
			return err
		} else {
			return GetTorrentList()
		}
	} else {
		goch.DisplayHeaderTable(goch.GetTableDataAsArrayWithHeaders(data, "//table[@id=\"torrent\"]", 0, 1))
		return nil
	}
}

func AddTorrent(filename string, magnet string) error {
	url := "https://upload.alldebrid.com/uploadtorrent.php"

	if uid, err := getUid(); err != nil {
		return err
	} else {
		form := curl.NewForm()
		form.Add("uid", uid)
		form.Add("domain", "https://www.alldebrid.com/torrent/")
		form.Add("magnet", magnet)
		if filename != "" {
			form.AddFile("uploadedfile", filename)
		}
		form.Add("submit", "Convert this torrent")

		if _, err := sendRequest(url, form, 302, false); err != nil {
			return err
		} else if filename != "" {
			fmt.Printf("%v correctly added to torrent queue\n", filename)
			return nil
		} else {
			fmt.Println("magnet correctly added to torrent queue")
			return nil
		}
	}
}

func RemoveTorrent(id string) error {
	url := "https://www.alldebrid.com/torrent/?action=remove&id=" + id

	if _, err := getUid(); err != nil {
		return err
	} else if data, err := sendRequest(url, nil, 302, false); err != nil && data != "" {
		return fmt.Errorf("ID %v not found in torrent queue", id)
	} else if err != nil {
		return err
	} else {
		fmt.Printf("ID %v correctly removed from torrent queue\n", id)
		return nil
	}
}

func DebridLink(link string) error {
	if url, filename, err := getDownloadLink(link); err != nil {
		return err
	} else {
		easy := curl.EasyInit()
		defer easy.Cleanup()

		started := int64(0)

		fp, _ := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		defer fp.Close()

		easy.Setopt(curl.OPT_URL, url)
		easy.Setopt(curl.OPT_VERBOSE, false)
		easy.Setopt(curl.OPT_NOPROGRESS, false)
		easy.Setopt(curl.OPT_PROGRESSFUNCTION, func(dltotal, dlnow, _, _ float64, _ interface{}) bool {
			if started == 0 {
				started = time.Now().Unix()
			}

			percentage := dlnow / dltotal * 100
			speed := dlnow / 1048576 / float64((time.Now().Unix() - started))

			fmt.Printf("Downloading %s: %3.2f%%, Speed: %.1fMB/s \r", filename, percentage, speed)

			return true
		})
		easy.Setopt(curl.OPT_WRITEFUNCTION, func(ptr []byte, userdata interface{}) bool {
			file := userdata.(*os.File)
			if _, err := file.Write(ptr); err != nil {
				return false
			}

			return true
		})
		easy.Setopt(curl.OPT_WRITEDATA, fp)

		if err := easy.Perform(); err != nil {
			fmt.Println(err.Error())
		}

		time.Sleep(1000000000) // wait gorotine

		fmt.Println()

		return nil
	}
}
