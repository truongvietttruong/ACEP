package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func fileExists(fpath string) bool {
	_, err := os.Stat(fpath)
	return err == nil
}

func appAction(c *cli.Context) error {
	inputFilepath := c.String("inputFilepath")
	outputFilepath := c.String("outputFilepath")
	invalidOutputFilepath := c.String("invalidOutputFilepath")
	delimiter := c.String("delimiter")
	loggingSteps := c.Int("loggingSteps")

	fmt.Printf("Read lines from %v\n", inputFilepath)

	inputFile, err := os.Open(inputFilepath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	inputScanner := bufio.NewScanner(inputFile)

	fmt.Printf("Start concatenating the lines into %v\n", outputFilepath)

	if fileExists(outputFilepath) {
		return fmt.Errorf("file already exists: %v", outputFilepath)
	}

	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)

	if fileExists(invalidOutputFilepath) {
		return fmt.Errorf("file already exists: %v", invalidOutputFilepath)
	}

	invalidOutputFile, err := os.Create(invalidOutputFilepath)
	if err != nil {
		return err
	}
	defer invalidOutputFile.Close()

	invalidWriter := bufio.NewWriter(invalidOutputFile)

	lineCount := 0
	for inputScanner.Scan() {
		if lineCount%loggingSteps == 0 {
			fmt.Printf("%v\r", lineCount/loggingSteps)
		}

		line := inputScanner.Text()

		tabSplits := strings.SplitN(line, "\t", 2)
		splits := strings.SplitN(tabSplits[0], delimiter, 2)
		if len(splits) != 2 {
			invalidWriter.WriteString(fmt.Sprintf("%v\t%v\n", tabSplits[0], tabSplits[1]))
			continue
		}

		writer.WriteString(fmt.Sprintf("%v\t%v\t%v\n", splits[0], splits[1], tabSplits[1]))

		lineCount++
	}
	if err := inputScanner.Err(); err != nil {
		return err
	}

	fmt.Println()

	writer.Flush()
	invalidWriter.Flush()

	return nil
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "inputFilepath",
				Aliases: []string{"i"},
			},
			&cli.StringFlag{
				Name:    "outputFilepath",
				Aliases: []string{"o"},
			},
			&cli.StringFlag{
				Name:    "invalidOutputFilepath",
				Aliases: []string{"v"},
			},
			&cli.StringFlag{
				Name:    "delimiter",
				Aliases: []string{"d"},
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
