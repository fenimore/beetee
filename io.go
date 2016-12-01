package main

import "os"
import "path/filepath"

// spawnFileWriter will spawn the goroutine which writes the file/files to disk. files might be empty.
func spawnFileWriter(name string, single bool, files []*TorrentFile) (chan *Piece, chan struct{}) {
	// TODO: Check if file/ Directories exist:
	//if _, err := os.Stat("/path/to/whatever"); err == nil {
	//        path/to/whatever exists
	//}
	writeSync.Add(len(d.Pieces))

	in := make(chan *Piece, FILE_WRITER_BUFSIZE)
	close := make(chan struct{})

	if single {
		f, err := os.Create(name)
		if err != nil {
			debugger.Println("Unable to create file")
		}
		defer f.Close()

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
		if _, err := os.Stat(name); err != nil {
			err = os.Mkdir(name, os.ModeDir|os.ModePerm)
			if err != nil {
				debugger.Println("Unable to make directory")
			}
		}

		go func() {
			createFiles(name, files)
			for {
				piece := <-in
				logger.Printf("Writing Data to Disk, Piece: %d", piece.index)
				// TODO: WriteToDisk
				writeMultipleFiles(piece, d.Torrent.Info.Files)
				writeSync.Done()
			}
		}()
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

func createFiles(name string, files []*TorrentFile) {
	for _, file := range files {
		if len(file.Path) < 2 {
			f, err := os.Create(filepath.Join(name, file.Path[0]))
			if err != nil {
				debugger.Println("Error creation file", file.Path)
			}
			defer f.Close()
		} else {
			// TODO: when there are sub directory
		}
	}
}

func writeMultipleFiles(piece *Piece, files []*TorrentFile) {

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
