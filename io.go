package main

import (
	"bufio"
	"errors"
	"os"
)

func (info *TorrentInfo) WriteData() error {
	//k	buffer := bufio.NewWriter(w io.Write)
	file, err := os.Create(info.Name)
	if err != nil {
		debugger.Println("File creation err: ", err)
		return err
	}
	writer := bufio.NewWriter(file)
	for idx, val := range Pieces {
		if !val.have {
			fi, err := file.Stat()
			if err != nil {
				debugger.Println(err)
			}
			debugger.Printf("File is %d bytes, out of Length: %d", fi.Size(), Torrent.Info.Length)
			debugger.Println("WHy Doesn't I have?", val)
			debugger.Println(idx)
			msg := string(idx) + " Is not had"
			return errors.New(msg)
		}
		writer.Write(val.data)
	}
	writer.Flush()
	logger.Println("Success Writing Data") // Not working?
	return nil
}

func (info *TorrentInfo) ContinuousWrite() error {
	fullFile := true
	for {
		file, err := os.Create(info.Name)
		if err != nil {
			debugger.Println("File creation err: ", err)
		}
		writer := bufio.NewWriter(file)
		for _, val := range Pieces {
			if val.have {
				//blank := make([]byte, info.PieceLength)
				//writer.Write(blank)
				//fullFile = false
				writer.Write(val.data)
			} else {
				fullFile = false
				if len(PieceQueue) < 1 && val.status == 0 {
					debugger.Println("Refilling Queue")
					PieceQueue <- val
				}

			}
		}
		fi, err := file.Stat()
		if err != nil {
			debugger.Println("error reading file", err)
		}
		if fi.Size() == info.Length {
			fullFile = true
		}
		writer.Flush()
		if fullFile {
			break
		}
	}
	logger.Println("Success Writing Data") // Not working?
	writeSync.Done()
	return nil
}

// TODO: Write iteratively onto desk
// TODO: Read the progress of blocks from disk
