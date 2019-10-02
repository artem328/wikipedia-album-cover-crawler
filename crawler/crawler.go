package crawler

import (
	"fmt"
	"log"
)

const UrlIndexPattern = "/w/index.php?title=Category:%d_albums"

type Crawler struct {
	grabber                *Grabber
	processor              Processor
	store                  *DetailsStorage
	urls                   chan string
	results                chan *HtmlPage
	targetDirectory        string
	maxConcurrentDownloads uint
}

func NewCrawler(targetDirectory string, maxConcurrentDownloads uint) *Crawler {
	parser := NewParser()
	return &Crawler{
		grabber: NewGrabber(),
		processor: NewChainProcessor([]Processor{
			NewListProcessor(parser),
			NewAlbumProcessor(parser),
			NewCoverProcessor(),
		}),
		store:                  NewDetailsStorage(),
		targetDirectory:        targetDirectory,
		maxConcurrentDownloads: maxConcurrentDownloads,
	}
}

func (c *Crawler) Start(years []int) {
	c.urls = make(chan string, 512)
	results := make(chan *HtmlPage, 512)

	go c.runGrabber(results)

	for _, year := range years {
		c.urls <- createUrl(fmt.Sprintf(UrlIndexPattern, year))
	}

	c.runProcessor(results)
}

func (c *Crawler) runGrabber(results chan *HtmlPage) {
	queue := make(chan struct{}, c.maxConcurrentDownloads)

	for url := range c.urls {
		log.Println("Grabbing url", url, "In queue", len(c.urls))
		queue <- struct{}{}
		go func(url string, results chan *HtmlPage, queue <-chan struct{}) {
			c.grabber.fetch(url, results)
			<-queue
		}(url, results, queue)
	}
}

func (c *Crawler) runProcessor(results <-chan *HtmlPage) {
	for result := range results {
		go c.processor.process(result, c)
	}
}
