package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

type Combolist struct {
	Email  string `parquet:"name=email,type=BYTE_ARRAY,convertedtype=UTF8"`
	POH    string `parquet:"name=poh,type=BYTE_ARRAY,convertedtype=UTF8"`
	Source string `parquet:"name=source,type=BYTE_ARRAY,convertedtype=UTF8"`
}

func convertTSVToParquet(inputFilepath string, outputFilepath string) error {
	fw, err := local.NewLocalFileWriter(outputFilepath)
	if err != nil {
		return err
	}

	pw, err := writer.NewParquetWriter(fw, new(Combolist), 3)
	if err != nil {
		return err
	}

	pw.RowGroupSize = 128 * 1024 * 1024
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	tsvFile, err := os.Open(inputFilepath)
	if err != nil {
		return err
	}
	defer tsvFile.Close()

	scanner := bufio.NewScanner(tsvFile)

	for scanner.Scan() {
		line := scanner.Text()
		splits := strings.Split(line, "\t")

		record := Combolist{
			Email:  splits[0],
			POH:    splits[1],
			Source: splits[2],
		}
		if err := pw.Write(record); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if err := pw.WriteStop(); err != nil {
		return err
	}

	return nil
}

func appAction(c *cli.Context) error {
	inputDir := c.String("inputDir")
	outputDir := c.String("outputDir")
	startChr := []rune(c.String("startChr"))[0]
	endChr := []rune(c.String("endChr"))[0]

	if err := os.MkdirAll(outputDir, 0666); err != nil {
		return err
	}

	for c := startChr; c <= endChr; c++ {
		fmt.Println(string(c))

		inputFilepath := filepath.Join(inputDir, string(c)+".txt")
		outputFilepath := filepath.Join(outputDir, string(c)+".parquet")

		if err := convertTSVToParquet(inputFilepath, outputFilepath); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "inputDir",
				Aliases: []string{"i"},
			},
			&cli.StringFlag{
				Name:    "outputDir",
				Aliases: []string{"o"},
			},
			&cli.StringFlag{
				Name:    "startChr",
				Aliases: []string{"s"},
				Value:   "a",
			},
			&cli.StringFlag{
				Name:    "endChr",
				Aliases: []string{"e"},
				Value:   "z",
			},
		},

		Action: appAction,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
