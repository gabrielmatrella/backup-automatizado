package main

import (
	"archive/zip"
	"io"
	"os"
	"testing"
)

func TestCompressing(t *testing.T) {
	w := createZipWriter("done.zip")
	defer w.Close()

	fileName := "teste-config.json"

	fileToZip, err := os.Open(fileName)
	defer fileToZip.Close()

	if err != nil {
		panic(err)
	}

	info, err := fileToZip.Stat()
	if err != nil {
		panic(err)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		panic(err)
	}

	header.Name = fileName
	header.Method = zip.Deflate

	fileWriter, err := w.CreateHeader(header)

	if err != nil {
		panic(err)
	}

	_, err = io.Copy(fileWriter, fileToZip)

	if err != nil {
		panic(err)
	}
}
