package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/cheggaaa/pb/v3"
	"github.com/urfave/cli/v2"
)

func writeToFile(writer *bufio.Writer, invalidWriter *bufio.Writer, inputFilepath string, delimiter string) error {
	fp, err := os.Open(inputFilepath)
	if err != nil {
		return err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, "\x00", "", -1)
		splits := strings.SplitN(line, delimiter, 2)
		if len(splits) != 2 {
			invalidWriter.WriteString(fmt.Sprintf("%v\t%v\n", line, inputFilepath))
			continue
		}

		writer.WriteString(fmt.Sprintf("%v\t%v\t%v\n", splits[0], splits[1], inputFilepath))
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	writer.Flush()
	invalidWriter.Flush()

	return nil
}

func appAction(c *cli.Context) error {
	inputRootDir := c.String("inputRootDir")
	outputFilepath := c.String("outputFilepath")
	invalidOutputFilepath := c.String("invalidOutputFilepath")
	errorOutputFilepath := c.String("errorOutputFilepath")
	delimiter := c.String("delimiter")

	globPattern := filepath.Join(inputRootDir, "**", "*.txt")
	inputFilepaths, err := doublestar.Glob(globPattern)
	if err != nil {
		return err
	}

	fmt.Printf("Start concatenating the lines into %v\n", outputFilepath)

	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)

	invalidOutputFile, err := os.Create(invalidOutputFilepath)
	if err != nil {
		return err
	}
	defer invalidOutputFile.Close()

	invalidWriter := bufio.NewWriter(invalidOutputFile)

	errorOutputFile, err := os.Create(errorOutputFilepath)
	if err != nil {
		return err
	}
	defer errorOutputFile.Close()

	errorWriter := bufio.NewWriter(errorOutputFile)

	bar := pb.StartNew(len(inputFilepaths))

	for _, inputFilepath := range inputFilepaths {
		bar.Increment()

		if err := writeToFile(writer, invalidWriter, inputFilepath, delimiter); err != nil {
			errorWriter.WriteString(fmt.Sprintln(inputFilepath))
			continue
		}
	}

	bar.Finish()

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
				Name:    "outputFilepath",
				Aliases: []string{"o"},
				Value:   "./concat.txt",
			},
			&cli.StringFlag{
				Name:    "invalidOutputFilepath",
				Aliases: []string{"v"},
				Value:   "./invalid.txt",
			},
			&cli.StringFlag{
				Name:    "errorOutputFilepath",
				Aliases: []string{"e"},
				Value:   "./errors.txt",
			},
			&cli.StringFlag{
				Name:    "delimiter",
				Aliases: []string{"d"},
			},
		},

		Action: appAction,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
