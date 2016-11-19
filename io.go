package main

import (
	"bufio"
	"errors"
	"os"
)

// WriteData writes the file to disk, if the pieces aren't all there
// it returns an error message, but rights all the same.
func (info *TorrentInfo) WriteData() error {
	//k	buffer := bufio.NewWriter(w io.Write)
	file, err := os.Create(info.Name)
	if err != nil {
		debugger.Println("File creation err: ", err)
		return err
	}
	writer := bufio.NewWriter(file)
	for idx, val := range Pieces {
		if !val.verified {
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

// NOTE DEPRECATED// ContinuousWrite writes even if pieces are missing.
// When the lenght matches up, or if all pieces are there,
// then it terminates and writes to disk.
/*func (info *TorrentInfo) ContinuousWrite() error {
	queueSync.Wait() // don't start writing until atleast the queue of requests is made.
	debugger.Println("Beginning ContinuousWrite")
	fullFile := true
	for {
		file, err := os.Create(info.Name)
		if err != nil {
			debugger.Println("File creation err: ", err)
		}
		writer := bufio.NewWriter(file)
		for _, val := range Pieces {
			if val.verified {
				//blank := make([]byte, info.PieceLength)
				//writer.Write(blank)
				//fullFile = false
				writer.Write(val.data)
			} else {
				fullFile = false
				//ndebugger.Println(len(PieceQueue), val.status, val.index)
				if len(PieceQueue) < 1 && val.verified {
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
*/
