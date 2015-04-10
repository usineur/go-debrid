package main

import (
	"errors"
	"github.com/usineur/go-debrid/alldebrid"
	"os"
)

func main() {
	fct := func() error {
		return errors.New("Usage: alldebrid [-d link | -t torrent | -m magnet | -r torrent_id | -l | -da | -ra]")
	}

	if args := len(os.Args); args == 2 && os.Args[1] == "-l" {
		fct = func() error { return alldebrid.Torrent("list") }
	} else if args == 2 && os.Args[1] == "-da" {
		fct = func() error { return alldebrid.Torrent("download_all") }
	} else if args == 2 && os.Args[1] == "-ra" {
		fct = func() error { return alldebrid.Torrent("remove_all") }
	} else if args == 3 && os.Args[1] == "-t" {
		fct = func() error { return alldebrid.AddTorrent(os.Args[2], "") }
	} else if args == 3 && os.Args[1] == "-m" {
		fct = func() error { return alldebrid.AddTorrent("", os.Args[2]) }
	} else if args == 3 && os.Args[1] == "-r" {
		fct = func() error { return alldebrid.Torrent("remove", os.Args[2]) }
	} else if args == 3 && os.Args[1] == "-d" {
		fct = func() error { return alldebrid.DebridLink(os.Args[2]) }
	}

	if err := fct(); err != nil {
		println(err.Error())
	}
}
