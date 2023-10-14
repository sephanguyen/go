package migrationsdata

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/scanner"
)

type MigratorConfig struct {
	fileInput        string
	folderOutput     string
	tableName        string
	batchSize        int
	totalConcurrency int
}

type IOutils interface {
	GetFoldersOnDir(filename string) ([]string, error)
	GetFileContent(filepath string) ([]byte, error)
	WriteFile(filename string, content []byte) error
}

type Migrator struct {
	MigratorConfig
	c       Converter
	IOUtils IOutils
}

type DataOutput struct {
	dataValid [][]string
	errors    [][]string
}

func InitMigrator(config MigratorConfig, ioutil IOutils) *Migrator {
	return &Migrator{
		MigratorConfig: config,
		IOUtils:        ioutil,
	}
}

func (s *Migrator) SetConverter(c Converter) {
	s.c = c
}

func (s *Migrator) worker(data <-chan DataOutput) <-chan string {
	out := make(chan string)
	index := 1
	if _, err := os.Stat(s.folderOutput); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(s.folderOutput, os.ModePerm)
		if err != nil {
			out <- fmt.Sprintf("create folder error %s", err)
		}
	}
	fWork := func(wg *sync.WaitGroup) {
		defer wg.Done()
		for data := range data {
			csvData := append([][]string{s.c.GetHeader()}, data.dataValid...)
			errors := data.errors
			dataConverted := exporter.ToCSV(csvData)
			dataErrors := exporter.ToCSV(errors)

			fileName := fmt.Sprintf("%s%d.csv", s.tableName, index)
			fileNameError := fmt.Sprintf("%s%d_error.csv", s.tableName, index)
			fileOut := fmt.Sprintf("%s/%s", s.folderOutput, fileName)
			fileOutError := fmt.Sprintf("%s/%s", s.folderOutput, fileNameError)

			index++

			if err := s.IOUtils.WriteFile(fileOut, dataConverted); err != nil {
				out <- fmt.Sprintf("write File error %s", err)
				return
			}
			if err := s.IOUtils.WriteFile(fileOutError, dataErrors); err != nil {
				out <- fmt.Sprintf("write File error %s", err)
				return
			}
			out <- fmt.Sprintf("Complete write file %s", fileName)
		}
	}
	go func() {
		wg := new(sync.WaitGroup)
		defer close(out)
		for i := 0; i < s.totalConcurrency; i++ {
			wg.Add(1)
			go fWork(wg)
		}
		wg.Wait()
	}()

	return out
}

func (s *Migrator) reader(ctx context.Context) (<-chan DataOutput, error) {
	data, err := s.IOUtils.GetFileContent(s.fileInput)
	if err != nil {
		return nil, err
	}
	out := make(chan DataOutput, s.totalConcurrency)
	sc := scanner.NewCSVScanner(bytes.NewReader(data))

	go func() {
		defer close(out)

		output := &DataOutput{
			dataValid: make([][]string, 0, s.batchSize),
			errors:    make([][]string, 0, s.batchSize),
		}

		for {
			scanned := sc.Scan()
			select {
			case <-ctx.Done():
				return
			default:

				if len(output.dataValid) == s.batchSize || len(output.errors) == s.batchSize || !scanned {
					out <- *output
					output = &DataOutput{
						dataValid: make([][]string, 0, s.batchSize),
						errors:    make([][]string, 0, s.batchSize),
					}
				}
				err := s.c.ValidationData(sc)
				if len(err) > 0 {
					output.errors = append(output.errors, err)
				} else {
					line := s.c.GetLineConverted(sc, orgID)
					output.dataValid = append(output.dataValid, line)
				}
			}
			if !scanned {
				return
			}
		}
	}()

	return out, nil
}

func (s *Migrator) Run(ctx context.Context, orgID string) error {
	chData, err := s.reader(ctx)
	if err != nil {
		return err
	}
	for result := range s.worker(chData) {
		fmt.Println(result)
	}
	fmt.Println("Done")
	return nil
}
