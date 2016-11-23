# beetee

Bittorrent Client implemented in Go **Work in Progress**. I have a blog post outlining the protocol in dialog format [here](http://another.workingagenda.com/blog/post/d1alog/).

> run beetee -file=linux.torrent

    beetee, commandline torrent application. Usage:
      -file string
            path to torrent file

====

# TODO:

- [ ] allow multiple file-torrents
- [ ] control peer flow/ask for more peers
- [x] parse pieces
- [x] put into pieces struct
- [x] Parse Have and BitField

## Downloading

- [x] ask peer for index
- [ ] find rarest blocks
- [x] only ask peer if they have it

### blocks

- [x] put that block into piece by index
- [x] concat blocks into data field

### Write to disk

- [x] manage blocks
- [x] write to disk
- [ ] write to disk gradually
- [ ] write and read when incomplete

## Uploading

- [x] run server
- [ ] construct bitfield
- [x] allow handshake
- [ ] parse request
- [ ] send blocks

====

Package Organisation:

`torrent`

> Torrent/meta/info structs and parse method.

`message`

> Handshake, individual message decoder and message sender.

`tracker`

> Tracker struct and Response method.

`server`

> Server struct and Listen method.

`peer`

> Peer struct, connect and Listen method. Also, the message decode switch for payloads is here. The peer constructor comes from parsing a TrackerResponse.

`piece`

> Piece/Block struct and piece parser from torrent info.

`io`

> Writing and reading to disk.

`manager`

> Manage peer and piece threads/channels/lists. Also, catch-all for now.
