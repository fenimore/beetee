[![Go Report Card](https://goreportcard.com/badge/github.com/polypmer/beetee)](https://goreportcard.com/report/github.com/polypmer/beetee) [![CircleCI](https://circleci.com/gh/polypmer/beetee.svg?style=shield)](https://circleci.com/gh/polypmer/beetee)

# beetee

Bittorrent Client implemented in Go **Work in Progress**. I have a blog post outlining the protocol in dialog format [here](http://another.workingagenda.com/blog/post/d1alog/).

> $ ./beetee -file=linux.torrent

    beetee, commandline torrent application. Usage:
      -file string
            path to torrent file
      -peers int
            max peer connections (default 30)
      -seed
            keep running after download completes


Thanks @kracekumar, @alex-segura, and @nschuc, fellow Recursers, for all your help :)

====

# Functionality:

- [x] allow multiple file-torrents
- [x] allow multiple dir-torrents
- [x] parse pieces
- [x] put into pieces struct
- [x] barse Have and BitField
- [x] begin Unit and Integration Tests
- [ ] control peer flow/ask for more peers

## Downloading

- [x] ask peer for index
- [x] only ask peer if they have it
- [ ] write test for last piece download
- [ ] ask for rarest blocks first

### blocks

- [x] put that block into piece by index
- [x] concat blocks into data field

### Write to disk

- [x] manage blocks
- [x] write to disk
- [x] write to disk gradually
- [ ] read when incomplete and put into pieces

## Uploading

- [x] run server
- [x] allow handshake
- [x] parse request
- [x] send blocks
- [ ] construct bitfield from pieces

====

Package Organisation:

`torrent`

> Torrent/meta/info structs and parse method. The torrent file provides the list of pieces.

`tracker`

> Tracker struct and Response method. The tracker provides a list of peers.

`message`

> Handshake, individual message decoder and message constructor.

`peer`

> Peer struct, connect and Listen method. Also, the message decode switch for payloads is here. This file is mostly IO for the peer sockets.

`server`

> Server struct and Listen method.

`piece`

> Piece/Block struct and piece parser from torrent info. Also piece validator and helper functions for parsing the last piece/blocks in a download.

`io`

> Writing and reading to disk.


For testing, here are the checksums for the torrent files provided in `torrents/` <small>Note: this Torrent file is for an older version of Arch; go to their website and get the newer torrent file/md5sum if you plan to use for installing onto your computer.</small>:

    17643c29e3c4609818f26becf76d29a3 > Ubuntu
    47672450bcda8acf0c8512bd5b543cc0 > Arch
