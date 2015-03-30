package main

import (
	"errors"
	"github.com/usineur/go-debrid/alldebrid"
	"os"
)

func main() {
	fct := func() error {
		return errors.New("Usage: alldebrid [-d link | -t torrent | -r torrent_id | -l]")
	}

	if args := len(os.Args); args == 2 && os.Args[1] == "-l" {
		fct = func() error { return alldebrid.GetTorrentList() }
	} else if args == 3 && os.Args[1] == "-t" {
		fct = func() error { return alldebrid.AddTorrent(os.Args[2]) }
	} else if args == 3 && os.Args[1] == "-r" {
		fct = func() error { return alldebrid.RemoveTorrent(os.Args[2]) }
	} else if args == 3 && os.Args[1] == "-d" {
		fct = func() error { return alldebrid.DebridLink(os.Args[2]) }
	}

	if err := fct(); err != nil {
		println(err.Error())
	}
}
