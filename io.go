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
	for _, val := range info.PieceList {
		//debugger.Println(idx)
		if !val.have {
			return errors.New("All pieces are not had")
		}
		writer.Write(val.data)
	}
	writer.Flush()
	logger.Println("Success Writing Data") // Not working?
	return nil
}
