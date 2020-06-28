package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/timshannon/bolthold"

	"github.com/WinPooh32/tracker-scraper/model"
	"github.com/WinPooh32/tracker-scraper/scraper"
)

func work(tCl *torrent.Client, scraper scraper.Scraper, proxy *url.URL, mdl *model.Torrent, result chan<- *model.Torrent) error {
	c, err := scraper.ScrapeContent(mdl.URL, proxy)
	if err != nil {
		return err
	}

	t, err := addMagnet(tCl, c.Magnet)
	if err != nil {
		return err
	}

	var ok bool
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	select {
	case <-ctx.Done():
		fmt.Println("Get info FAIL: ", t.Name())
	case <-t.GotInfo():
		ok = true
	}

	if ok {
		var files []metainfo.FileInfo
		for _, f := range t.Files() {
			files = append(files, f.FileInfo())
		}

		mdl.Description = c.Value
		mdl.Magnet = c.Magnet

		mdl.Files = files
		mdl.Peers = uint32(t.Stats().TotalPeers)

		mdl.Complete = true
	}

	mdl.Try++

	select {
	case result <- mdl:
	case <-ctx.Done():
	}

	t.Drop()
	cancel()

	return nil
}

func pipe(store *bolthold.Store, result <-chan *model.Torrent) {
	for mdl := range result {
		err := store.Upsert(mdl.URL.String(), mdl)
		if err != nil {
			panic(err)
		}
	}
}

func addMagnet(cl *torrent.Client, magnet metainfo.Magnet) (*torrent.Torrent, error) {
	var err error
	var t *torrent.Torrent

	var spec = &torrent.TorrentSpec{
		Trackers:    [][]string{magnet.Trackers},
		DisplayName: magnet.DisplayName,
		InfoHash:    magnet.InfoHash,
	}

	t, _, err = cl.AddTorrentSpec(spec)
	if err != nil {
		return nil, fmt.Errorf("torrent add magnet: %w", err)
	}

	return t, nil
}
