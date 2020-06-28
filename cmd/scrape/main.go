package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/timshannon/bolthold"
	"golang.org/x/sync/semaphore"

	"github.com/WinPooh32/tracker-scraper/model"
	"github.com/WinPooh32/tracker-scraper/scraper/rutracker"
)

func main() {
	var succeed uint32

	store, err := bolthold.Open("bolthold.db", 0666, nil)
	if err != nil {
		panic(err)
	}

	proxy, _ := url.Parse("http://127.0.0.1:8118")

	cfg := torrent.NewDefaultClientConfig()
	cfg.HTTPProxy = http.ProxyURL(proxy)
	cfg.DisableIPv6 = true

	cl, err := torrent.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	r := &rutracker.Rutracker{}

	for i := 0; i < 5; i++ {
		fmt.Printf("page %d\n", i)

		u, _ := url.Parse(fmt.Sprintf("https://rutracker.org/forum/viewforum.php?f=1727&start=%d", i*50))

		lst, err := r.ScrapeCatalog(u, proxy)
		if err != nil {
			panic(err)
		}

		for _, URL := range lst {
			var u = URL

			var mdl = &model.Torrent{
				URL:      &u.URL,
				Title:    u.Title,
				Complete: false,
			}

			err := store.Insert(mdl.URL.String(), mdl)
			if err != nil && err != bolthold.ErrKeyExists {
				panic(err)
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	result := make(chan *model.Torrent, 50)

	ctx := context.TODO()
	sema := semaphore.NewWeighted(25)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		err = store.ForEach(bolthold.Where("Complete").Not().Eq(true), func(record *model.Torrent) error {
			if record.Complete {
				return nil
			}

			err = sema.Acquire(ctx, 1)
			if err != nil {
				panic(err)
			}

			wg.Add(1)

			go func() {
				defer wg.Done()
				defer sema.Release(1)
				if err := work(cl, r, proxy, record, result); err == nil {
					atomic.AddUint32(&succeed, 1)
					fmt.Println("DONE: ", succeed)
				}
			}()

			return nil
		})
		if err != nil {
			panic(err)
		}
	}()

	go pipe(store, result)

	wg.Wait()
	close(result)

	fmt.Println("DONE ALL")
}
