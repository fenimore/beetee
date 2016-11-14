# beetee
Bittorrent Client implemented in Go


[x] parse pieces
[x] put into pieces struct
[?] put index into pieces struct
[ ] Parse Have and BitField

## Downloading
[x] ask peer for index // not a big deal
### blocks
[ ] put that block into piece by index
[ ] concat blocks into data field
### Write to disk
[ ] manage blocks
[ ] write to disk

## Uploading
