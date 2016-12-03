# beetee

Bittorrent Client implemented in Go **Work in Progress**. I have a blog post outlining the protocol in dialog format [here](http://another.workingagenda.com/blog/post/d1alog/).

> run $ ./beetee -file=linux.torrent

    beetee, commandline torrent application. Usage:
      -file string
            path to torrent file
      -seed
            keep running after download completes

Thanks @kracekumar and @alex-segura, fellow Recursers, for all your help :)

====

# TODO:

- [x] allow multiple file-torrents
- [x] allow multiple dir-torrents
- [ ] control peer flow/ask for more peers
- [x] parse pieces
- [x] put into pieces struct
- [x] Parse Have and BitField
- [x] Begin Unit and Integration Tests

## Downloading

- [x] ask peer for index
- [ ] ask for rarest blocks first
- [x] only ask peer if they have it

### blocks

- [x] put that block into piece by index
- [x] concat blocks into data field

### Write to disk

- [x] manage blocks
- [x] write to disk
- [x] write to disk gradually
- [ ] read when incomplete

## Uploading

- [x] run server
- [ ] construct bitfield
- [x] allow handshake
- [x] parse request
- [x] send blocks

====

Package Organisation:

`torrent`

> Torrent/meta/info structs and parse method.


`tracker`

> Tracker struct and Response method.

`message`

> Handshake, individual message decoder and message constructor.

`peer`

> Peer struct, connect and Listen method. Also, the message decode switch for payloads is here. The peer constructor comes from parsing a TrackerResponse. This file is mostly IO for peer sockets.

`server`

> Server struct and Listen method.

`piece`

> Piece/Block struct and piece parser from torrent info. Also piece validator and helper functions for parsing the last piece/blocks in a download.

`io`

> Writing and reading to disk.


For testing, here are the checksums for the torrent files provided in `torrents/`:

    17643c29e3c4609818f26becf76d29a3 > Ubuntu
    47672450bcda8acf0c8512bd5b543cc0 > Arch
