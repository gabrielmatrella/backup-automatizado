package main

import (
	"archive/zip"
	"io"
	"os"
	"testing"
)

func TestCompressing(t *testing.T) {
	w := createZipWriter(createZipFile("done.zip"))
	defer w.Close()

	fileName := "config.json"

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

func TestGenerateFileName(t *testing.T) {
	result := generateZipFileName()

	if result == "" {
		t.Fail()
	} else {
		t.Log(result)
	}
}

func TestGenerateFileName2(t *testing.T) {
	os.Open("C:\\Users\\Gabriel\\Documents\\Nekrotus\\ORDEM_DAS_POCOES.TXT")

	select {}
}
