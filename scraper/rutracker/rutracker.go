package rutracker

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/anacrolix/torrent/metainfo"

	"github.com/WinPooh32/tracker-scraper/http"
	"github.com/WinPooh32/tracker-scraper/scraper"

	"github.com/PuerkitoBio/goquery"
)

var ErrPostBodyNotFound = errors.New("post body not found")
var ErrMagnetNotFound = errors.New("magnet link not found")

type status string

const (
	statusNone        status = ""
	statusDup                = "tor-dup"
	statusConsumed           = "tor-consumed"
	statusApproved           = "tor-approved"
	statusNotApproved        = "tor-not-approved"
)

const host = "https://rutracker.org"

func toStatus(class string) status {
	var scan = bufio.NewScanner(strings.NewReader(class))
	scan.Split(bufio.ScanWords)

	for scan.Scan() {
		var txt = scan.Text()

		switch txt {
		case statusDup:
			fallthrough
		case statusConsumed:
			fallthrough
		case statusApproved:
			fallthrough
		case statusNotApproved:
			return status(txt)
		}
	}

	return statusNone
}

type Rutracker struct{}

func (r *Rutracker) ScrapeCatalog(u, proxy *url.URL) ([]url.URL, error) {
	var err error

	var body []byte
	var list []url.URL
	var doc *goquery.Document

	body, err = http.Get(u, proxy)
	if err != nil {
		return nil, err
	}

	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	doc.Find("td > .torTopic").Each(func(i int, s *goquery.Selection) {
		var st status
		var link string

		var icon = s.Find(".tor-icon")
		if icon.Length() == 0 {
			return
		}

		if class, ok := icon.Attr("class"); ok {
			if st = toStatus(class); st == statusNone {
				return
			}
		}

		var a = s.Find("a")
		if a.Length() == 0 {
			return
		}

		var ok bool
		link, ok = a.Attr("href")
		if !ok {
			return
		}

		var u *url.URL
		u, err = url.Parse(host + "/forum/" + link)
		if err != nil {
			return
		}

		list = append(list, *u)
	})

	return list, nil
}

func (r *Rutracker) ScrapeContent(u, proxy *url.URL) (*scraper.Content, error) {
	var err error

	var htm string
	var mgt metainfo.Magnet

	var body []byte
	var doc *goquery.Document

	body, err = http.Get(u, proxy)
	if err != nil {
		return nil, err
	}

	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var post = doc.Find(".post_body")
	if post.Length() == 0 {
		return nil, ErrPostBodyNotFound
	}

	htm, err = post.First().Html()
	if err != nil {
		return nil, err
	}

	var link = doc.Find(".magnet-link")
	if link.Length() == 0 {
		return nil, ErrMagnetNotFound
	}

	mgt, err = metainfo.ParseMagnetURI(link.AttrOr("href", ""))
	if err != nil {
		return nil, fmt.Errorf("parse magnet uri: %w", err)
	}

	return &scraper.Content{
		Value:  htm,
		Magnet: mgt,
	}, nil
}
