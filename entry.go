package main

import (
	"archive/zip"
	"compress/flate"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/onatm/clockwerk"
	"github.com/tkanos/gonfig"
)

const CONFIG_FILE_NAME = "config.json"

var globalConfig Configuration

var PATH_SEPARATOR string

func main() {
	// configure path separator
	if runtime.GOOS == "windows" {
		PATH_SEPARATOR = "\\"
	} else {
		PATH_SEPARATOR = "/"
	}

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
				root = strings.TrimRight(root, PATH_SEPARATOR)

				// reusable after first write folder
				var firstZipFilePath string
				var zipFileName string = generateZipFileName()

				for i, wPath := range globalConfig.WriterPaths {
					currWriteFilePath := wPath + PATH_SEPARATOR + zipFileName

					if i == 0 {
						zipFile := createZipFile(currWriteFilePath)
						w := createZipWriter(zipFile)

						lastPathIndex := strings.LastIndex(root, PATH_SEPARATOR)
						err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
							// skip root folder
							if path == root {
								return nil
							}

							// append filtered file
							if !info.IsDir() && isValidExtension(info.Name()[strings.LastIndex(info.Name(), ".")+1:len(info.Name())]) {
								log.Println("Scan:", path)
								pathToSave := path[lastPathIndex+1:]
								return addFileToZip(path, pathToSave, w)
							}

							return nil
						})

						if err != nil {
							fmt.Println(err)
						}

						w.Close()
						zipFile.Close()

						firstZipFilePath = currWriteFilePath
					} else {
						// copy zip from first backup location
						copyFile(currWriteFilePath, firstZipFilePath)
					}

					log.Println("Copy to", wPath+PATH_SEPARATOR+zipFileName)
				}
			}

		}
	}
}

func addFileToZip(filePath string, fileInZipPath string, w *zip.Writer) error {
	f, err := os.Open(filePath)

	if err != nil {
		fmt.Println(err)
	}

	fileInfo, err := os.Stat(filePath)

	if err != nil {
		return err
	}

	h, err := zip.FileInfoHeader(fileInfo)

	if err != nil {
		return err
	}

	h.Name = fileInZipPath
	h.Method = zip.Deflate

	fW, err := w.CreateHeader(h)
	if err != nil {
		return err
	}

	_, err = io.Copy(fW, f)
	if err != nil {
		return err
	}

	return nil
}

func copyFile(dest, src string) bool {
	destFile, err := os.Create(dest)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	srcFile, err := os.Open(src)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	_, err = io.Copy(destFile, srcFile)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	destFile.Close()
	srcFile.Close()

	return true
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
	return time.Now().Format("Backup_20060102150405") + ".zip"
}

func createZipFile(zipFileName string) *os.File {
	newZipFile, _ := os.Create(zipFileName)

	return newZipFile
}

func startJobs() {
	var job BackupJob

	// Run first job
	go job.Run()
	c := clockwerk.New()
	c.Every(time.Duration(globalConfig.Frequency) * time.Minute).Do(job)
	c.Start()
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
		return false
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
