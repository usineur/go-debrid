package alldebrid

import (
	"encoding/json"
	"fmt"
	"github.com/andelf/go-curl"
	"html"
	"os"
	"sort"
	"strings"
	"time"
)

type service struct {
	Link      string
	Host      interface{} // could be a string or a boolean
	Filename  string
	Icon      string
	Streaming interface{} // could be []string or map[string]inteface{}
	Nb        int
	Error     string
	Filesize  string
}

func getDownloadLink(link string) (string, string, bool, error) {
	fields := map[string]string{
		"json": "true",
		"link": link,
	}
	var s service

	if res, _, err := sendRequest("/service.php", fields, nil); err != nil {
		return "", "", false, err
	} else if res == "login" {
		if err := getCookie(); err != nil {
			return "", "", false, err
		}

		return getDownloadLink(link)
	} else if err := json.Unmarshal([]byte(res), &s); err != nil {
		return "", "", false, err
	} else if s.Error != "" {
		return "", "", false, fmt.Errorf(s.Error)
	} else if s.Link != "" {
		return s.Link, s.Filename, false, nil
	} else {
		link, suffix := getStreamLink(s.Streaming)
		return link, html.UnescapeString(s.Filename) + suffix, true, nil
	}
}

func getStreamLink(streaming interface{}) (string, string) {
	sLinks := streaming.(map[string]interface{})

	var description []string
	choice := -1
	suffix := ".mp4"
	err := fmt.Errorf("")

	for i := range sLinks {
		description = append(description, i)
	}
	sort.Strings(description)

	fmt.Println("Only stream links are available. Please choose one entry.")
	for i, j := range description {
		fmt.Printf("\t%v - %v\n", i, j)
	}

	for err != nil {
		if choice, err = getChoice(len(description)); err != nil || choice == -1 {
			choice = -1
			err = fmt.Errorf("Invalid choice")
			fmt.Println(err)
		} else {
			err = nil
		}
	}

	if strings.HasPrefix(description[choice], "audio") {
		suffix = ".mp3"
	}

	return sLinks[description[choice]].(string), suffix
}

func DebridLink(link string) error {
	if url, filename, stream, err := getDownloadLink(link); err != nil {
		return err
	} else {
		fp, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		defer fp.Close()

		if stream {
			fmt.Printf("Downloading of \"%v\"\n", filename)
			if err := netcat(fp, url); err != nil {
				return err
			} else {
				fmt.Printf("\nDownload finished\n")
				return nil
			}
		}

		easy := curl.EasyInit()
		defer easy.Cleanup()

		started := int64(0)

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
			fmt.Printf("\n%v", err.Error())
		}

		time.Sleep(1000000000) // wait gorotine

		fmt.Println()

		return nil
	}
}
