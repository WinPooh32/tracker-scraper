package http

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html/charset"
)

const getTimeout = 30 * time.Second

func Get(url, proxy *url.URL) ([]byte, error) {
	var transport = &http.Transport{
		Proxy:               http.ProxyURL(proxy),
		TLSHandshakeTimeout: getTimeout,
	}

	var client = &http.Client{
		Transport: transport,
		Timeout:   getTimeout,
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", url.Host)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:69.0) Gecko/20100101 Firefox/69.0")
	req.Header.Set("Accept", "text/html,*/*;q=0.1")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var reader io.Reader

	if contentType, ok := resp.Header["Content-Type"]; ok && len(contentType) > 0 {
		var _, contentEncoding, _ = charset.DetermineEncoding(nil, contentType[0])

		reader, err = charset.NewReader(resp.Body, contentEncoding)
		if err != nil {
			return nil, err
		}
	} else {
		reader = resp.Body
	}

	return ioutil.ReadAll(reader)
}
