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
		// NOTE: Because it's created
		// No need for perm flags?
		if err != nil {
			debugger.Println("Unable to create file")
		}
		f.Close()

		go func() {
			for {
				f, err = os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0777)
				if err != nil {
					debugger.Println("ERror opening file: ", err)
				}
				select {
				case piece := <-in:
					logger.Printf("Writing Data to Disk, Piece: %d", piece.index)
					n, err := f.WriteAt(piece.data, int64(piece.index)*piece.size)
					if err != nil {
						debugger.Printf("Error Writing %d bytes: %s", n, err)
					}
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
				writeMultipleFiles(piece, name, d.Torrent.Info.Files)
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
	// TODO: Create when there are sub directories
	for _, file := range files {
		if len(file.Path) < 2 {
			if _, err := os.Stat(filepath.Join(name, file.Path[0])); err != nil {
				// if file doesn't exist
				f, err := os.Create(filepath.Join(name, file.Path[0]))
				if err != nil {
					debugger.Println("Error creating file", file.Path)
				}
				defer f.Close()
			}
		} else {
			var path string
			filename := file.Path[len(file.Path)-1]
			path = filepath.Join(name)
			for _, val := range file.Path[:len(file.Path)-1] { // leave out the file name
				path = filepath.Join(path, val)
			}

			if _, err := os.Stat(path); err != nil {
				os.MkdirAll(path, os.ModePerm)
			}

			if _, err := os.Stat(filepath.Join(path, filename)); err != nil {
				// if file doesn't exist
				f, err := os.Create(filepath.Join(path, filename))
				if err != nil {
					debugger.Println("Error creating file", file.Path)
				}
				defer f.Close()
			}
		}
	}
}

func writeMultipleFiles(piece *Piece, name string, files []*TorrentFile) {
	// NOTE: Not fully supported:
	// NOTE: Won't work if piece extends past two files
	for idx, file := range files {
		// These bounds are relative to the total Piece list
		pieceLower := int64(piece.index) * piece.size            // 0    or 16
		pieceUpper := int64(piece.index)*piece.size + piece.size // 16 kb   32
		// fileLower := file.PreceedingTotal
		fileUpper := file.PreceedingTotal + file.Length
		if pieceLower > fileUpper {
			continue // Wrong File
		}

		path := filepath.Join(name)
		for _, val := range file.Path {
			path = filepath.Join(path, val)
		}

		f, err := os.OpenFile(path,
			os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			debugger.Println("Error Opening/Writing Piece %d to file %s",
				piece.index, file.Path)
		}

		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			debugger.Println("Err Getting file stats", err)
		}

		if pieceUpper <= fileUpper {
			fileOffset := int64(piece.index)*piece.size - file.PreceedingTotal
			n, err := f.WriteAt(piece.data, fileOffset)
			if err != nil {
				debugger.Println("Write Error", n, err)
			}

			f.Close()
			break
		} else { // Piece Extends to multple files

			//carrySize := pieceUpper - fi.Size()
			//fileUpper // How much of the piece overflows
			carrySize := pieceUpper - fileUpper
			carry := piece.data[:carrySize]
			data := piece.data[carrySize:]
			fileOffset := file.Length - int64(len(data))
			debugger.Println("")
			debugger.Println("Upper Bounds", fileUpper, "<", pieceUpper)
			debugger.Println("Lower Bounds", file.PreceedingTotal, "<", pieceLower)
			debugger.Println("offset     :", fileOffset)
			debugger.Println("Add it toget", fileOffset+carrySize, file.Length)
			debugger.Println("Sizes:     :", fi.Size(), file.Length)
			debugger.Println("Add size   :", fi.Size()+int64(len(data)))
			//debugger.Println("Other stats:", file.PreceedingTotal -
			debugger.Printf("Data size: %d, Carry size: %d", len(data), len(carry))
			debugger.Println("")

			n, err := f.WriteAt(data, fileOffset)
			if err != nil {
				debugger.Println("err %s || writing amount: %d", err, n)
			}
			fi, err := f.Stat()
			if err != nil {
				debugger.Println("Err Getting file stats", err)
			}
			if fi.Size() != file.Length {
				debugger.Printf("File is not the write size: Got %d, Expected %d",
					fi.Size(), file.Length)
				debugger.Printf("Data size: %d, Carry size: %d", len(data), len(carry))
			}

			// The next file
			path = filepath.Join(name)
			for _, val := range files[idx+1].Path {
				path = filepath.Join(path, val)
			}

			nxtFile, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0777)
			if err != nil {
				debugger.Println("Error opening Next file")
			}
			defer nxtFile.Close()

			n, err = nxtFile.WriteAt(carry, 0)
			if err != nil {
				debugger.Println("Write Error", n, err)
			}
		}
	}

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
