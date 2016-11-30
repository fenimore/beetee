package main

import "os"

func spawnFileWriter(f *os.File) (chan *Piece, chan struct{}) {

	writeSync.Add(len(d.Pieces))

	in := make(chan *Piece, FILE_WRITER_BUFSIZE)
	close := make(chan struct{})
	go func() {
		for {
			select {
			case piece := <-in:
				logger.Printf("Writing Data to Disk, Piece: %d", piece.index)
				f.WriteAt(piece.data, int64(piece.index)*piece.size)
				writeSync.Done()
			case <-close:
				f.Close()
			}
		}
	}()
	return in, close
}

func checkFileSize(filename string) (int64, error) {
	// TODO: If it is only one file
	file, err := os.OpenFile(filename, os.O_RDONLY, os.ModeTemporary)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	fi, _ := file.Stat()
	return fi.Size(), nil
}
