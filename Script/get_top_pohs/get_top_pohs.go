package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/urfave/cli/v2"
)

func addCounts(counts map[string]int, inputFilepath string) error {
	fp, err := os.Open(inputFilepath)
	if err != nil {
		return err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		line := scanner.Text()
		splits := strings.Split(line, "\t")

		key := splits[0]

		count, err := strconv.Atoi(splits[1])
		if err != nil {
			return err
		}

		if curCount, exists := counts[key]; exists {
			counts[key] = curCount + count
		} else {
			counts[key] = count
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func sortCounts(counts map[string]int) ([]string, []int) {
	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range counts {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool { return ss[i].Value > ss[j].Value })

	keys := make([]string, 0)
	values := make([]int, 0)
	for _, kv := range ss {
		keys = append(keys, kv.Key)
		values = append(values, kv.Value)
	}

	return keys, values
}

func appAction(c *cli.Context) error {
	inputDir := c.String("inputDir")
	outputFilepath := c.String("outputFilepath")

	globPattern := filepath.Join(inputDir, "*.txt")
	inputFilepaths, err := doublestar.Glob(globPattern)
	if err != nil {
		return err
	}

	counts := make(map[string]int)
	for _, inputFilepath := range inputFilepaths {
		fmt.Println(inputFilepath)

		if err := addCounts(counts, inputFilepath); err != nil {
			return err
		}
	}

	sortedKeys, _ := sortCounts(counts)

	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for i := 0; i < len(sortedKeys); i++ {
		k := sortedKeys[i]
		writer.WriteString(k + "\n")
	}
	writer.Flush()

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
				Name:    "outputFilepath",
				Aliases: []string{"o"},
			},
		},

		Action: appAction,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
