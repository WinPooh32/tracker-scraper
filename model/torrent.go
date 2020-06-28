package model

import (
	"net/url"

	"github.com/anacrolix/torrent/metainfo"
)

type Torrent struct {
	URL *url.URL

	Title       string
	Description string

	Magnet metainfo.Magnet

	Files []metainfo.FileInfo
	Peers uint32

	Try      uint32
	Complete bool
}
