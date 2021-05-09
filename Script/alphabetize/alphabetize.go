package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"unicode"

	"github.com/urfave/cli/v2"
)

func appAction(c *cli.Context) error {
	inputFilepath := c.String("inputFilepath")
	outputDir := c.String("outputDir")
	startChr := []rune(c.String("startChr"))[0]
	endChr := []rune(c.String("endChr"))[0]
	loggingSteps := c.Int("loggingSteps")

	inputFile, err := os.Open(inputFilepath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	inputScanner := bufio.NewScanner(inputFile)

	if err := os.MkdirAll(outputDir, 0666); err != nil {
		return err
	}

	//First, create files and writers corresponding to all the processed characters
	writers := make(map[rune]*bufio.Writer)

	for c := startChr; c <= endChr; c++ {
		outputFilepath := filepath.Join(outputDir, string(c)+".txt")

		outputFile, err := os.OpenFile(outputFilepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		defer outputFile.Close()

		writer := bufio.NewWriter(outputFile)
		writers[c] = writer
	}

	//Then, scan the target file and sort by the initial character of each line
	lineCount := 0

	fmt.Printf("Logging Steps: %v\n", loggingSteps)

	for inputScanner.Scan() {
		if lineCount%loggingSteps == 0 {
			fmt.Printf("%v\r", int(lineCount/loggingSteps))
		}

		line := inputScanner.Text()

		lineCount += 1

		firstChr := []rune(line)[0]
		firstChr = unicode.ToLower(firstChr)

		writer, exists := writers[firstChr]
		if exists {
			writer.WriteString(fmt.Sprintln(line))
		}
	}
	if err := inputScanner.Err(); err != nil {
		return err
	}

	for _, writer := range writers {
		writer.Flush()
	}

	return nil
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "inputFilepath",
				Aliases: []string{"i"},
				Value:   "./concat.txt",
			},
			&cli.StringFlag{
				Name:    "outputDir",
				Aliases: []string{"o"},
				Value:   "./Alphabetized",
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
			&cli.IntFlag{
				Name:    "loggingSteps",
				Aliases: []string{"l"},
				Value:   1000000,
			},
		},

		Action: appAction,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
