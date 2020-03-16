package main

import (
	"archive/zip"
	"compress/flate"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/onatm/clockwerk"
	"github.com/tkanos/gonfig"
)

const CONFIG_FILE_NAME = "config.json"

var globalConfig Configuration

type Configuration struct {
	BackupPaths []string
	WriterPaths []string
	Extensions  []string
	Frequency   int32
	DaysOfWeek  string // 1-SEG, 2-TER, 3-QUA, 4-QUI, 5-SEX, 6-SAB, 7-DOM *-TODOS
}

type BackupJob struct{}

func (b BackupJob) Run() {
	// Backup worker
	currentWeekday := int(time.Now().Weekday())
	daysConfig := globalConfig.DaysOfWeek
	isJobDay := strings.Contains(daysConfig, "*") || strings.Contains(daysConfig, strconv.Itoa(currentWeekday))

	if isJobDay {
		if writerPathsExist() {
			for _, root := range globalConfig.BackupPaths {
				zipFile := createZipFile(generateZipFileName())
				w := createZipWriter(zipFile)

				defer zipFile.Close()
				defer w.Close()

				err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
					// skip root folder
					if path == root {
						return nil
					}

					// append filtered file
					if info.IsDir() {

					} else if isValidExtension(info.Name()[strings.LastIndex(info.Name(), ".")+1 : len(info.Name())]) {
						header, err := zip.FileInfoHeader(info)

						if err != nil {
							fmt.Println(err)
						}

						header.Name = path
						header.Method = zip.Deflate

						fileWriter, _ := w.CreateHeader(header)

						f, err := os.Open(path)

						if err != nil {
							fmt.Println(err)
						}

						io.Copy(fileWriter, f)
					}

					return nil
				})

				if err != nil {
					fmt.Println(err)
				}

				w.Close()
			}
		}
	}
}

func isValidExtension(fileExtension string) bool {
	for _, val := range globalConfig.Extensions {
		if strings.Contains(val, "*") || strings.Contains(strings.ToUpper(val), strings.ToUpper(fileExtension)) {
			return true
		}
	}

	return false
}

func generateZipFileName() string {
	return time.Now().Format("Backup_200601021504") + ".zip"
}

func main() {
	if configFileExists() {
		globalConfig = Configuration{}
		err := gonfig.GetConf(CONFIG_FILE_NAME, &globalConfig)

		if err != nil {
			panic(err)
		}

		// start jobs in separete routine
		startJobs()

		// lock application to keep it running
		for {
			select {}
		}
	}
}

func startJobs() {
	var job BackupJob

	// Run first job
	go job.Run()

	c := clockwerk.New()
	c.Every(time.Duration(globalConfig.Frequency) * time.Minute).Do(job)
	c.Start()
}

func createZipFile(zipFileName string) *os.File {
	newZipFile, _ := os.Create(zipFileName)

	return newZipFile
}

func createZipWriter(zipFile *os.File) *zip.Writer {
	// Create new zip archive
	w := zip.NewWriter(zipFile)
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
		BackupPaths: []string{"mypath1/data", "mypath2/data"},
		WriterPaths: []string{"C:/MyBackupPath"},
		Extensions:  []string{"TXT"},
		DaysOfWeek:  "1 2 3",
		Frequency:   20, // MINUTES
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

func writerPathsExist() bool {
	pathsToWrite := globalConfig.WriterPaths

	for _, val := range pathsToWrite {
		_, err := os.Stat(val)

		if os.IsNotExist(err) {
			err := os.Mkdir(val, os.ModeDir)
			if os.IsPermission(err) {
				fmt.Println("Acesso negado ao criar diretorios de escrita")
			} else if err != nil {
				fmt.Println(err.Error())
			} else {
				return true
			}
		}
	}

	return true
}
