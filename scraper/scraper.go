package scraper

import (
	"net/url"

	"github.com/anacrolix/torrent/metainfo"
)

type Page struct {
	Title string
	URL   url.URL
}

type Content struct {
	Value  string
	Magnet metainfo.Magnet
}

type Scraper interface {
	ScrapeCatalog(u, proxy *url.URL) ([]Page, error)
	ScrapeContent(u, proxy *url.URL) (*Content, error)
}
