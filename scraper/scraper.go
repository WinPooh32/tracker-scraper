package scraper

import (
	"net/url"

	"github.com/anacrolix/torrent/metainfo"
)

type Content struct {
	Value  string
	Magnet metainfo.Magnet
}

type Scraper interface {
	ScrapeCatalog(u, proxy *url.URL) ([]url.URL, error)
	ScrapeContent(u, proxy *url.URL) (*Content, error)
}
