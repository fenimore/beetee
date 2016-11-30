package main

import "os"

func spawnFileWriter(f *os.File) (chan *Piece, chan struct{}) {
	in := make(chan *Piece, FILE_WRITER_BUFSIZE)
	close := make(chan struct{})
	go func() {
		for {
			select {
			case piece := <-in:
				logger.Printf("Writing Data to Disk, Piece: %d", piece.index)
				f.WriteAt(piece.data, int64(piece.index)*piece.size)
			case <-close:
				f.Close()
			}
		}
	}()
	return in, close
}
