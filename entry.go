package main

import (
	"archive/zip"
	"compress/flate"
	"encoding/json"
	"io"
	"os"

	"github.com/tkanos/gonfig"
)

type Configuration struct {
	BackupPaths    []string
	WriterPaths    []string
	Extensions     []string
	Frequency      int8
	StartRunningAt string
	StopRunningAt  string
	DaysOfWeek     string // SEG, TER, QUA, QUI, SEX, SAB, DOM
}

const CONFIG_FILE_NAME = "teste-config.json"

func main() {
	if configFileExists() {
		configuration := Configuration{}
		err := gonfig.GetConf(CONFIG_FILE_NAME, &configuration)

		if err != nil {
			panic(err)
		}

	}
}

func createZipWriter(zipFile string) *zip.Writer {
	newZipFile, _ := os.Create(zipFile)
	// Create new zip archive
	w := zip.NewWriter(newZipFile)
	// Register a custom Deflate compressor
	w.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestSpeed)
	})

	return w
}

func configFileExists() bool {
	_, err := os.Stat(CONFIG_FILE_NAME)

	if os.IsNotExist(err) {
		return createDefaultConfigFile()
	} else if err != nil {
		return false
	}

	return true
}

func createDefaultConfigFile() bool {
	file, createFileError := os.Create(CONFIG_FILE_NAME)

	configuration := Configuration{
		BackupPaths:    []string{"mypath1/data", "mypath2/data"},
		WriterPaths:    []string{"C:/MyBackupPath"},
		Extensions:     []string{"TXT"},
		DaysOfWeek:     "SEG, TER, QUA, QUI, SEX, SAB, DOM",
		Frequency:      4,
		StartRunningAt: "08:00",
		StopRunningAt:  "17:00",
	}

	jsonBytes, err := json.Marshal(&configuration)

	if err == nil {
		file.WriteString(string(jsonBytes))
	}

	if createFileError != nil {
		return false
	} else {
		return true
	}
}
