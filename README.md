go-debrid
===========

Tool written in Go to interact with Alldebrid API

How to use
----------
* Debrid a link supported by Alldebrid
```
        --debrid, -d    <link>
```
* Add a torrent/magnet, can be used with extra paramaters "```--split```" to file files in parts of 1 Gb, and/or "```--disable-quick```" to disable quick search of existing torrents in system
```
        --torrent, -t   <torrent file> [--split, -s|--disable-quick, -q]
        --magnet,  -m   <magnet link>  [--split, -s|--disable-quick, -q]
```
* List torrents in queue
```
        --list, -l
```
* Remove a torrent (tip: use "```--list```" and check column "ID" to get the entry to remove)
```
        --remove, -r    <torrent id>
```
* Remove all torrents in queue
```
        --remove-all, -ra
```
* Download all finished torrents
```
        --download-all, -da
```

Credits
=======
- Alldebrid: http://www.alldebrid.com

