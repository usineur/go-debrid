package alldebrid

import (
	"fmt"
	"github.com/usineur/go-debrid/utils"
	"regexp"
)

func RemainingTime() error {
	if res, eff, err := sendRequest("/account/", nil, nil); err != nil {
		return err
	} else if eff == host+"/" {
		if err := getCookie(); err != nil {
			return err
		} else {
			return RemainingTime()
		}
	} else if msg, err := utils.GetContent(res, "//*[@id=\"account_top_right\"]/div[4]"); err != nil {
		return err
	} else {
		re := regexp.MustCompile("[ ]+")
		fmt.Println(re.ReplaceAllString(msg, " "))
		return nil
	}
}
