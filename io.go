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
					// TODO: doesn't work for last piece duhh!
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
				logger.Printf("Writing Data to Disk, Piece: %d out of: %d", piece.index,
					len(d.Pieces))
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
	switch mode := fi.Mode(); {
	case mode.IsDir():
		var size int64
		err = filepath.Walk(
			filename,
			func(
				_ string,
				info os.FileInfo,
				err error) error {
				if !info.IsDir() {
					size += info.Size()
				}
				return err
			})
		return size, nil
	case mode.IsRegular():
		return fi.Size(), nil
	default:
		return fi.Size(), nil
	}
}

func createFiles(name string, files []*TorrentFile) {
	// TODO: Create when there are sub directories
	for _, file := range files {
		logger.Println("Creating:", file.Path, file.Length)
		// construct path
		var path string
		filename := file.Path[len(file.Path)-1]
		path = filepath.Join(name)
		for _, val := range file.Path[:len(file.Path)-1] { // leave out the file name
			path = filepath.Join(path, val)
		}
		// Create directory
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
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func abs(a int64) int64 {
	if a > 0 {
		return a
	}
	return -a
}

func writeMultipleFiles(piece *Piece, name string, files []*TorrentFile) {
	for _, file := range files {
		// TODO: pass in TorrentInf
		ok, data, offset := pieceInFile(piece, file, d.Torrent.Info.PieceLength)
		if !ok {
			continue
		}
		// Get file path
		path := filepath.Join(name)
		for _, val := range file.Path {
			path = filepath.Join(path, val)
		}
		f, err := os.OpenFile(path,
			os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			debugger.Println("Error Write %d to file %s",
				piece.index, file.Path)
		}
		defer f.Close()

		n, err := f.WriteAt(data, offset)
		if err != nil {
			debugger.Println("Write Error", n, err)
		}
	}
}

// pieceInFile returns the data to be written on a file, and it's offset
func oldPieceInFile(piece *Piece, file *TorrentFile) (bool, []byte, int64) {
	// TODO: Piece size for last piece?
	pieceLower := int64(piece.index) * piece.size
	pieceUpper := int64(piece.index+1) * piece.size
	fileUpper := file.PreceedingTotal + file.Length
	if pieceLower > fileUpper || pieceUpper < file.PreceedingTotal {
		return false, nil, 0
	}

	offset := max(0, pieceLower-file.PreceedingTotal)
	lower := abs(min(0, pieceLower-file.PreceedingTotal))

	upper := min(file.PreceedingTotal+file.Length, pieceUpper-file.PreceedingTotal+file.Length)

	return true, piece.data[lower:upper], offset
}

// pieceInFile returns the data to be written on a file, and it's offset
func pieceInFile(piece *Piece, file *TorrentFile, pieceSize int64) (bool, []byte, int64) {
	pieceLower := int64(piece.index) * pieceSize
	pieceUpper := int64(piece.index+1) * pieceSize
	fileUpper := file.PreceedingTotal + file.Length
	if pieceLower > fileUpper || pieceUpper < file.PreceedingTotal {
		// NOTE: Some files aren't in the 'write' space
		return false, nil, 0
	}

	offset := max(0, pieceLower-file.PreceedingTotal)
	lower := abs(min(0, pieceLower-file.PreceedingTotal))
	upper := min(file.Length-offset+lower, piece.size)
	if upper == lower {
		return false, nil, 0
	}
	//logger.Println(lower, upper)
	return true, piece.data[lower:upper], offset
}
