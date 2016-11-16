# beetee
Bittorrent Client implemented in Go


package organization for sanity:

## torrent
torrent/meta/info structs and parse method
## message
Handshake, Message Decoder and Message Senders
## tracker
Tracker struct and Response method...
## peer
Peer struct, connect and Listen Method
## piece
Piece/Block struct and piece parser from torrent info
## io
writing to disk


# TODO:
[x] parse pieces
[x] put into pieces struct
[?] put index into pieces struct
[ ] Parse Have and BitField

## Downloading
[x] ask peer for index // not a big deal
### blocks
[x] put that block into piece by index
[x] concat blocks into data field
### Write to disk
[x] manage blocks
[x] write to disk

## Uploading
