package main

import "os"

/*

 */

func spawnFileWriter(name string, single bool) (chan *Piece, chan struct{}) {
	in := make(chan *Piece, FILE_WRITER_BUFSIZE)
	close := make(chan struct{})
	if single {
		f, err := os.Create(name)
		if err != nil {
			debugger.Println("Unable to create file")
		}

		writeSync.Add(len(d.Pieces))

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
	} else {
		// write multiple files
	}
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

func multipleWrite() {
	/// single file lenght
	// lowerBound := index*size -size
	// upperBound := index*size
	// for file in files:
	// if file.size is > lowerBound and < upperBound:
	// perform spliting check
	// len%piecesize if not zero, then the remainder goes to the next file
	//
}
