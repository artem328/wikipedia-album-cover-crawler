package main

import (
	"github.com/artem328/wikipedia-album-cover-crawler/crawler"
	"log"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]

	if len(args) < 3 {
		log.Fatal("Invalid argument count")
	}

	targetDirectory := args[0]

	maxConcurrentDownloads, err := strconv.Atoi(args[1])

	if err != nil {
		log.Fatal("failed to parse max concurrent downloads argument", err)
	}

	yearsRaw := args[2:]
	years := make([]int, 0)
	for _, yearRaw := range yearsRaw {
		year, err := strconv.Atoi(yearRaw)
		if err != nil {
			log.Fatal("Year is not a number")
		}

		years = append(years, year)
	}

	c := crawler.NewCrawler(targetDirectory, uint(maxConcurrentDownloads))

	c.Start(years)
}
