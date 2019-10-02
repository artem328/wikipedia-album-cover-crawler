package crawler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Processor interface {
	supports(page *HtmlPage) bool
	process(page *HtmlPage, crawler *Crawler)
}

type ListProcessor struct {
	sync.Mutex
	parser *Parser
}

func NewListProcessor(parser *Parser) *ListProcessor {
	return &ListProcessor{parser: parser}
}

func (*ListProcessor) supports(page *HtmlPage) bool {
	return strings.Contains(page.url, "index.php?title=Category:")
}

func (p *ListProcessor) process(page *HtmlPage, crawler *Crawler) {
	result := p.parser.parseListItems(page.body)

	if result == nil {
		return
	}

	for _, u := range result.urls {
		crawler.urls <- u
	}

	if result.nextPageUrl != "" {
		crawler.urls <- result.nextPageUrl
	}
}

type AlbumProcessor struct {
	parser *Parser
}

func NewAlbumProcessor(parser *Parser) *AlbumProcessor {
	return &AlbumProcessor{parser: parser}
}

func (*AlbumProcessor) supports(page *HtmlPage) bool {
	return strings.Contains(page.url, "/wiki/")
}

func (p *AlbumProcessor) process(page *HtmlPage, crawler *Crawler) {
	details := p.parser.parseAlbumDetails(page.body)

	if details == nil {
		return
	}

	crawler.store.Add(details)
	crawler.urls <- details.coverUrl
}

type CoverProcessor struct {
	alphaRegexp *regexp.Regexp
}

func NewCoverProcessor() *CoverProcessor {
	return &CoverProcessor{
		alphaRegexp: regexp.MustCompile("(i?)^[[:alpha:]]"),
	}
}

func (*CoverProcessor) supports(page *HtmlPage) bool {
	return strings.Contains(page.url, "upload.wikimedia")
}

func (p *CoverProcessor) process(page *HtmlPage, crawler *Crawler) {
	details := crawler.store.DetailsForUrl(page.url)

	if details == nil {
		return
	}
	defer crawler.store.Remove(details)

	filename := p.generateName(details, crawler.targetDirectory)
	directory := path.Dir(filename)

	err := os.MkdirAll(directory, 0777)

	if err != nil {
		log.Println("Error when saving file", err)
		return
	}

	err = ioutil.WriteFile(filename, page.body, 0644)

	if err != nil {
		log.Println("Error when saving file", err)
		return
	}

	log.Println("Downloaded an image to", filename)
}

func (p *CoverProcessor) generateName(details *AlbumDetails, targetDirectory string) string {
	var year string
	if details.year == 0 {
		year = "other"
	} else {
		year = strconv.FormatInt(int64(details.year), 10)
	}

	var first string
	if p.alphaRegexp.MatchString(details.name) {
		first = strings.ToUpper(details.name[:1])
	} else {
		first = "other"
	}

	var extension string
	filePath := strings.Split(details.coverUrl, "/")
	filename := filePath[len(filePath)-1]
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		extension = parts[len(parts)-1]
	} else {
		extension = "jpg"
	}

	return fmt.Sprintf("%s/%s/%s/%s.%s",
		targetDirectory,
		year,
		first,
		p.sanitizeName(fmt.Sprintf("%s - %s", details.artist, details.name)),
		extension,
	)
}

func (*CoverProcessor) sanitizeName(name string) string {
	return url.QueryEscape(name)
}

type ChainProcessor struct {
	processors []Processor
}

func (*ChainProcessor) supports(page *HtmlPage) bool {
	return true
}

func (p *ChainProcessor) process(page *HtmlPage, crawler *Crawler) {
	for _, processor := range p.processors {
		if !processor.supports(page) {
			continue
		}

		processor.process(page, crawler)
		break
	}
}

func NewChainProcessor(processors []Processor) *ChainProcessor {
	return &ChainProcessor{processors: processors}
}
