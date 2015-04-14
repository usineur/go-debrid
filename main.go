package main

import (
	"errors"
	"github.com/usineur/go-debrid/alldebrid"
	"os"
)

func main() {
	args := len(os.Args)

	fct := func() error {
		return errors.New("Usage: go-debrid [OPTIONS]...")
	}

	if args > 1 {
		switch os.Args[1] {
		case "-l", "--list":
			if args == 2 {
				fct = func() error { return alldebrid.Torrent("list") }
			}
			break

		case "-da", "--download-all":
			if args == 2 {
				fct = func() error { return alldebrid.Torrent("download_all") }
			}
			break

		case "-ra", "--remove-all":
			if args == 2 {
				fct = func() error { return alldebrid.Torrent("remove_all") }
			}
			break

		case "-r", "--remove":
			if args == 3 {
				fct = func() error { return alldebrid.Torrent("remove", os.Args[2]) }
			}
			break

		case "-t", "--torrent", "-m", "--magnet":
			split := false
			quick := false
			valid := true

			if args < 3 {
				break
			} else if args == 4 {
				split = splitOption(os.Args[3])
				quick = disableQuickOption(os.Args[3])
				valid = split || quick
			} else if args == 5 {
				split = splitOption(os.Args[3]) || splitOption(os.Args[4])
				quick = disableQuickOption(os.Args[3]) || disableQuickOption(os.Args[4])
				valid = split && quick
			}

			if !valid {
				break
			} else if os.Args[1][1] == 't' {
				fct = func() error { return alldebrid.AddTorrent(os.Args[2], "", split, !quick) }
			} else {
				fct = func() error { return alldebrid.AddTorrent("", os.Args[2], split, !quick) }
			}
			break

		case "-d", "--debrid":
			if args == 3 {
				fct = func() error { return alldebrid.DebridLink(os.Args[2]) }
			}
			break

		default:
			break
		}
	}

	if err := fct(); err != nil {
		println(err.Error())
	}
}

func splitOption(value string) bool {
	return value == "--split" || value == "-s"
}

func disableQuickOption(value string) bool {
	return value == "--disable-quick" || value == "-q"
}
