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
	for _, val := range Pieces {
		//dbugger.Println(idx)
		if !val.have {
			//debugger.Println(Pieces[idx-1])
			//debugger.Println("WHy Doesn't I have?", val.
			debugger.Println("WHy Doesn't I have?", val)
			msg := string(val.index) + " Is not had"
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
