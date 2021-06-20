package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/urfave/cli/v2"
)

func alphabetize(
	inputFilepath string,
	delimiter string,
	writers map[rune]*bufio.Writer,
	invalidWriters map[rune]*bufio.Writer) error {
	fp, err := os.Open(inputFilepath)
	if err != nil {
		return err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, "\x00", "", -1)

		if len(line) == 0 {
			continue
		}

		firstChr := []rune(line)[0]
		firstChr = unicode.ToLower(firstChr)

		if _, exists := writers[firstChr]; !exists {
			continue
		}

		splits := strings.SplitN(line, delimiter, 2)
		if len(splits) != 2 {
			invalidWriter := invalidWriters[firstChr]
			invalidWriter.WriteString(fmt.Sprintf("%v\t%v\n", line, inputFilepath))

			continue
		}

		writer := writers[firstChr]
		writer.WriteString(fmt.Sprintf("%v\t%v\t%v\n", splits[0], splits[1], inputFilepath))
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for _, writer := range writers {
		writer.Flush()
	}
	for _, invalidWriter := range invalidWriters {
		invalidWriter.Flush()
	}

	return nil
}

func appAction(c *cli.Context) error {
	inputRootDir := c.String("inputRootDir")
	outputDir := c.String("outputDir")
	delimiter := c.String("delimiter")
	startChr := []rune(c.String("startChr"))[0]
	endChr := []rune(c.String("endChr"))[0]

	//Enumerate target files
	globPattern := filepath.Join(inputRootDir, "**", "*.txt")
	inputFilepaths, err := doublestar.Glob(globPattern)
	if err != nil {
		return err
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

	//Create a file to output errors
	errorOutputFilepath := filepath.Join(outputDir, "errors.txt")

	errorOutputFile, err := os.OpenFile(errorOutputFilepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer errorOutputFile.Close()

	errorWriter := bufio.NewWriter(errorOutputFile)

	//Scan the target files and sort by the initial character of each line
	for inputFileIdx, inputFilepath := range inputFilepaths {
		fmt.Printf("%v/%v\r", inputFileIdx, len(inputFilepaths))

		if err := alphabetize(inputFilepath, delimiter, writers, invalidWriters); err != nil {
			errorWriter.WriteString(fmt.Sprintln(inputFilepath))
			continue
		}
	}

	errorWriter.Flush()

	return nil
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "inputRootDir",
				Aliases: []string{"i"},
				Value:   "./Data",
			},
			&cli.StringFlag{
				Name:    "outputDir",
				Aliases: []string{"o"},
				Value:   "./Alphabetized",
			},
			&cli.StringFlag{
				Name:    "delimiter",
				Aliases: []string{"d"},
				Value:   ":",
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
