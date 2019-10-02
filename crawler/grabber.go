package crawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Grabber struct {
	client *http.Client
}

type HtmlPage struct {
	url  string
	body []byte
}

func NewGrabber() *Grabber {
	return &Grabber{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func NewHtmlPage(url string, body []byte) *HtmlPage {
	return &HtmlPage{url: url, body: body}
}

func (g *Grabber) fetch(url string, result chan<- *HtmlPage) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)

	if err != nil {
		fmt.Println("Error occurred while creating request", err)
		return
	}

	res, err := g.client.Do(req)

	if err != nil {
		fmt.Println("Error occurred while requesting", url, err)
		return
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			fmt.Println("Error occurred while closing response body", err)
		}
	}()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error occurred while reading body", err)
		return
	}

	result <- NewHtmlPage(url, body)
}
