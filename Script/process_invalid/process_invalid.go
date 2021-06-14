package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

func splitLines(
	inputFilepath string,
	delimiter string,
	writer *bufio.Writer,
	invalidWriter *bufio.Writer) error {
	fp, err := os.Open(inputFilepath)
	if err != nil {
		return err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		line := scanner.Text()
		//line = strings.Replace(line, "\x00", "", -1)

		if len(line) == 0 {
			continue
		}

		splits := strings.Split(line, "\t")
		combo := splits[0]
		origFilepath := splits[1]

		comboSplits := strings.SplitN(combo, delimiter, 2)
		if len(comboSplits) != 2 {
			invalidWriter.WriteString(fmt.Sprintln(line))
			continue
		}

		writer.WriteString(fmt.Sprintf("%v\t%v\t%v\n", comboSplits[0], comboSplits[1], origFilepath))
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	writer.Flush()
	invalidWriter.Flush()

	return nil
}

func appAction(c *cli.Context) error {
	inputDir := c.String("inputDir")
	outputDir := c.String("outputDir")
	delimiter := c.String("delimiter")
	startChr := []rune(c.String("startChr"))[0]
	endChr := []rune(c.String("endChr"))[0]

	if delimiter == "" {
		return errors.New("delimiter must be specified")
	}

	if err := os.MkdirAll(outputDir, 0666); err != nil {
		return err
	}

	//Create output files and writers corresponding to all the processed characters
	writers := make(map[rune]*bufio.Writer)
	invalidWriters := make(map[rune]*bufio.Writer)

	for c := startChr; c <= endChr; c++ {
		outputFilepath := filepath.Join(outputDir, string(c)+".txt")

		outputFile, err := os.OpenFile(outputFilepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		defer outputFile.Close()

		writer := bufio.NewWriter(outputFile)
		writers[c] = writer

		invalidOutputFilepah := filepath.Join(outputDir, "invalid_"+string(c)+".txt")

		invalidOutputFile, err := os.OpenFile(invalidOutputFilepah, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		defer invalidOutputFile.Close()

		invalidWriter := bufio.NewWriter(invalidOutputFile)
		invalidWriters[c] = invalidWriter
	}

	//Scan the target files
	for c := startChr; c <= endChr; c++ {
		inputFilepath := filepath.Join(inputDir, string(c)+".txt")
		fmt.Println(inputFilepath)

		if err := splitLines(inputFilepath, delimiter, writers[c], invalidWriters[c]); err != nil {
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
				Name:    "delimiter",
				Aliases: []string{"d"},
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
