package alldebrid

import (
	"encoding/json"
	"fmt"
	"github.com/andelf/go-curl"
	"github.com/usineur/goch"
	"os"
	"regexp"
	"strings"
	"time"
)

const host = "https://www.alldebrid.com"

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

func sendRequest(path string, data map[string]string, form interface{}) (string, string, error) {
	easy := curl.EasyInit()
	defer easy.Cleanup()

	doc := ""
	url := ""

	if url = host + path + goch.PrepareFields(data); form != nil {
		url = strings.Replace(url, "www", "upload", -1)
		easy.Setopt(curl.OPT_HTTPPOST, form)
	}

	easy.Setopt(curl.OPT_URL, url)
	easy.Setopt(curl.OPT_COOKIEFILE, cookie)
	easy.Setopt(curl.OPT_COOKIEJAR, cookie)
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
	os.Remove(cookie)

	id, pass := getCredentials()
	fields := map[string]string{
		"action":         "login",
		"login_login":    id,
		"login_password": pass,
	}

	if res, eff, err := sendRequest("/register/", fields, nil); err != nil {
		return err
	} else if eff != host+"/" {
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
	fields := map[string]string{
		"json": "true",
		"link": link,
	}
	var s service

	if res, _, err := sendRequest("/service.php", fields, nil); err != nil {
		return "", "", err
	} else if res == "login" {
		if err := getCookie(); err != nil {
			return "", "", err
		} else {
			return getDownloadLink(link)
		}
	} else if err := json.Unmarshal([]byte(res), &s); err != nil {
		return "", "", err
	} else if s.Error != "" {
		return "", "", fmt.Errorf(s.Error)
	} else {
		return s.Link, s.Filename, nil
	}
}

func GetTorrentList() error {
	if res, eff, err := sendRequest("/torrent/", nil, nil); err != nil {
		return err
	} else if eff == host+"/" {
		if err := getCookie(); err != nil {
			return err
		} else {
			return GetTorrentList()
		}
	} else {
		goch.DisplayHeaderTable(goch.GetTableDataAsArrayWithHeaders(res, "//table[@id=\"torrent\"]", 0, 1))
		return nil
	}
}

func AddTorrent(filename string, magnet string) error {
	if uid, err := getUid(); err != nil {
		return err
	} else {
		path := "/torrent/"

		form := curl.NewForm()
		form.Add("uid", uid)
		form.Add("domain", host+path)
		form.Add("magnet", magnet)
		if filename != "" {
			form.AddFile("uploadedfile", filename)
		}
		form.Add("submit", "Convert this torrent")

		if res, eff, err := sendRequest("/uploadtorrent.php", nil, form); err != nil {
			return err
		} else if res == "Bad uploaded files" {
			return fmt.Errorf(res)
		} else if pattern, err := regexp.Compile(host + path + "\\?error=(.*)"); err != nil {
			return err
		} else if matches := pattern.FindStringSubmatch(eff); len(matches) == 2 {
			switch matches[1] {
			case "":
				return fmt.Errorf("Alldebrid internal problem: retry")

			case "1":
				return fmt.Errorf("Problem in magnet link")

			case "2":
				return fmt.Errorf("This doesn't seem to be a magnet link.")

			case "3":
				return fmt.Errorf("You have to be premium.")

			case "4":
				return fmt.Errorf("Download already finished.")

			case "5":
				return fmt.Errorf("The torrent that you try to download is too big, can't be bigger than 60 Go.")

			default:
				return fmt.Errorf("Unsupported torrent error %v", matches[1])
			}
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
	path := "/torrent/"
	fields := map[string]string{
		"action": "remove",
		"id":     id,
	}

	if _, eff, err := sendRequest(path, fields, nil); err != nil {
		return err
	} else if eff == host+"/" {
		if err := getCookie(); err != nil {
			return err
		} else {
			return RemoveTorrent(id)
		}
	} else if eff != host+path {
		return fmt.Errorf("ID %v not found in torrent queue", id)
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
