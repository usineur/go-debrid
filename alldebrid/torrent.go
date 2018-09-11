package alldebrid

import (
	"fmt"
	"github.com/andelf/go-curl"
	"github.com/anisus/eol"
	"github.com/usineur/goch"
	"regexp"
	"strings"
)

func getUid() (string, error) {
	if contents, err := getFileContent(cookie); err != nil {
		if err := getCookie(); err != nil {
			return "", err
		} else {
			return getUid()
		}
	} else if pattern, err := regexp.Compile(".*\tuid\t(.*)" + eol.OS); err != nil {
		return "", err
	} else if matches := pattern.FindStringSubmatch(contents); len(matches) != 2 {
		return "", fmt.Errorf("Expected cookie \"uid\" not found\n")
	} else {
		return matches[1], nil
	}
}

func Torrent(params ...string) error {
	path := "/torrent/"

	var fields map[string]string
	if len(params) == 2 {
		fields = map[string]string{
			"action": params[0],
			"id":     params[1],
		}
	}

	if res, eff, err := sendRequest(path, fields, nil); err != nil {
		return err
	} else if eff == host+"/" {
		if err := getCookie(); err != nil {
			return err
		} else {
			return Torrent(params...)
		}
	} else if label, values, err := goch.GetTableDataAsArrayWithHeaders(res, "//table[@id=\"torrent\"]", 0, 1); err != nil {
		return err
	} else {
		switch params[0] {
		case "list":
			goch.DisplayHeaderTable(label, values, err)
			return nil

		case "download_all":
			for i, _ := range values {
				for _, link := range strings.Split(values[i][10], ",;,") {
					if link == "Pending" {
						fmt.Printf("ID %v is not yet downloadable\n", values[i][1])
					} else if link != "" {
						if err := DebridLink(link, ""); err != nil {
							fmt.Println(err.Error())
						}
					}
				}
			}

			if len(values) == 0 {
				return fmt.Errorf("Empty torrent list: nothing to download")
			} else {
				return nil
			}

		case "remove":
			if id := fields["id"]; eff != host+path {
				return fmt.Errorf("ID %v not found in torrent queue", id)
			} else {
				fmt.Printf("ID %v correctly removed from torrent queue\n", id)
				return nil
			}

		case "remove_all":
			for i, _ := range values {
				if err := Torrent("remove", values[i][1]); err != nil {
					fmt.Println(err.Error())
				}
			}

			if len(values) == 0 {
				return fmt.Errorf("Empty torrent list: nothing to remove")
			} else {
				return nil
			}

		default:
			return fmt.Errorf("Action \"%v\" not supported", params)
		}
	}
}

func AddTorrent(filename string, magnet string, split bool, quick bool) error {
	if uid, err := getUid(); err != nil {
		return err
	} else {
		path := "/torrent/"

		form := curl.NewForm()

		form.Add("uid", uid)
		form.Add("domain", host+path)
		form.Add("magnet", magnet)
		if filename != "" {
			form.AddFile("files[]", filename)
		}
		form.Add("splitfile", btos(split))
		form.Add("quick", btos(quick))
		form.Add("submit", "Convert this torrent")

		if res, eff, err := sendRequest("/uploadtorrent.php", nil, form); err != nil {
			return err
		} else if res == "Bad uploaded files" {
			return fmt.Errorf(res)
		} else if res == "Invalid cookie." {
			return fmt.Errorf(res)
		} else if pattern, err := regexp.Compile(host + path + "\\?error=(.*)"); err != nil {
			return err
		} else if matches := pattern.FindStringSubmatch(eff); len(matches) == 2 {
			switch matches[1] {
			case "":
				return fmt.Errorf("Alldebrid internal problem: retry")

			default:
				if msg, err := goch.GetContent(res, "//div[@style=\"color:red;text-align:center;\"]"); err != nil {
					return err
				} else {
					return fmt.Errorf(msg)
				}
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
