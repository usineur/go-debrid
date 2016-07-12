package alldebrid

import (
	"fmt"
	"github.com/usineur/goch"
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
	} else if msg, err := goch.GetContent(res, "//*[@id=\"account_top_right\"]/div[4]"); err != nil {
		return err
	} else {
		re := regexp.MustCompile("[ ]+")
		fmt.Println(re.ReplaceAllString(msg, " "))
		return nil
	}
}
