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

// TODO: Write iteratively onto desk
// TODO: Read the progress of blocks from disk
