package crawler

import "net/url"

const Host = "en.wikipedia.org"
const Scheme = "https"

func createUrl(rawUrl string) string {
	u, err := url.Parse(rawUrl)

	if err != nil {
		return rawUrl
	}

	if u.Host == "" {
		u.Host = Host
	}

	if u.Scheme == "" {
		u.Scheme = Scheme
	}

	return u.String()
}
