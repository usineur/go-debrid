package alldebrid

import (
	"encoding/json"
	"fmt"
	"github.com/andelf/go-curl"
	"os"
	"time"
)

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
