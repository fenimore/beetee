# beetee

Bittorrent Client implemented in Go **Work in Progress**

File Organisation:

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

## manager

manage peer and piece threads/channels/lists

====

# TODO:

- [x] parse pieces
- [x] put into pieces struct
- [ ] Parse Have and BitField

## Downloading

- [x] ask peer for index // not a big deal
- [ ] find rarest blocks
- [ ] only ask peer if they have it

### blocks

- [x] put that block into piece by index
- [x] concat blocks into data field

### Write to disk

- [x] manage blocks
- [x] write to disk
- [ ] write and read when incomplete

## Uploading

- [ ] allow handshake
- [ ] parse request
- [ ] send blocks
