package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Item struct {
	Start  int64
	Length int64
	Name   string
}

type ArchiveInfo struct {
	Items []Item
	Path  string
}

func compressArchive(archivePath, outPath string) bool {
	dupe, err := os.Create(outPath)
	if err != nil {
		log.Printf("Unable to open %v for reading\r\n", outPath)
		return false
	}
	defer dupe.Close()
	gzipPntr := gzip.NewWriter(dupe)
	defer gzipPntr.Close()

	basePntr, err := os.Open(archivePath)
	if err != nil {
		log.Printf("Unable to open %v for writing\r\n", archivePath)
		return false
	}
	defer basePntr.Close()

	io.Copy(gzipPntr, basePntr)

	return true
}

func mergeIntoBaseArchive(baseArchive ArchiveInfo, basedir string, contents []string, file string) bool {
	// tar pntr for copy
	dupe, err := os.Create(file)
	if err != nil {
		log.Printf("Unable to open %v for reading\r\n", file)
		return false
	}
	defer dupe.Close()

	tw := tar.NewWriter(dupe)
	defer tw.Close()

	basePntr, err := os.Open(baseArchive.Path)
	if err != nil {
		log.Printf("Unable to open archive %v for appending\r\n", baseArchive.Path)
		return false
	}
	defer basePntr.Close()

	written, err := io.Copy(dupe, basePntr)
	if written == 0 {
		log.Printf("Warning: Did not write anything from %v to %v\r\n", baseArchive.Path, file)
	}

	if err != nil {
		log.Printf("Copy failed: \r\n", err)
		return false
	}

	// bump to the end
	dupe.Seek(-2<<9, os.SEEK_END)

	// insert
	for _, item := range contents {
		res := writeFileToArchive(dupe, tw, item, basedir)
		if res == nil {
			log.Printf("Unable to add %v to new archive\r\n", item)
			return false
		}
	}

	return true
}

func createBaseArchive(basedir string, contents []string, file string) *ArchiveInfo {
	tarPntr, err := os.Create(file)
	if err != nil {
		log.Printf("Unable to open base archive %v\r\n", file)
		return nil
	}
	defer tarPntr.Close()

	tw := tar.NewWriter(tarPntr)
	defer tw.Close()
	total := len(contents)

	a := ArchiveInfo{Path: file}

	for index, file := range contents {
		item := writeFileToArchive(tarPntr, tw, file, basedir)
		if item == nil {
			log.Printf("Failed to add %v to base archive.\r\n", file)
			return nil
		}
		fmt.Printf("\rArchiving %v / %v", index+1, total)
		a.Items = append(a.Items, *item)
	}
	fmt.Println()

	return &a
}

func writeFileToArchive(tarPntr *os.File, tw *tar.Writer, file string, basedir string) *Item {
	curPos, err := tarPntr.Seek(0, 1)
	if err != nil {
		log.Println("Unable to determine current position")
		return nil
	}
	stat, err := os.Stat(file)
	if err != nil {
		log.Printf("Unable to query file %v\r\n", file)
		return nil
	}

	hdr := &tar.Header{
		Name:    strings.Replace(file, basedir, "", 1),
		Size:    stat.Size(),
		Mode:    775,
		ModTime: stat.ModTime(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		log.Printf("Unable to write TAR header for %v\r\n", hdr.Name)
		return nil
	}

	filePntr, err := os.Open(file)
	if err != nil {
		log.Printf("Unable to open %v for reading\r\n", hdr.Name)
		return nil
	}
	defer filePntr.Close()

	// read in chunks for memory
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := filePntr.Read(buf)
		if err != nil && err != io.EOF {
			log.Printf("Unable to open %v for reading\r\n", hdr.Name)
			return nil
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := tw.Write(buf[:n]); err != nil {
			log.Printf("Unable to write chunk for %v\r\n", hdr.Name)
			return nil
		}
	}

	endPos, err := tarPntr.Seek(0, 1)
	if err != nil {
		log.Println("Unable to determine end position")
		return nil
	}

	return &Item{Start: curPos, Length: (endPos - curPos), Name: hdr.Name}
}
