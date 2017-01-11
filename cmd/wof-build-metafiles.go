package main

import (
	"flag"
	"fmt"
	_ "github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"github.com/whosonfirst/go-whosonfirst-uri"
	_ "io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"
)

func main() {

	procs := flag.Int("processes", runtime.NumCPU()*2, "The number of concurrent processes to use")
	repo := flag.String("repo", "", "...")
	limit := flag.Int("open-filehandles", 512, "...")

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	abs_repo, err := filepath.Abs(*repo)

	if err != nil {
		log.Fatal(err)
	}

	info, err := os.Stat(abs_repo)

	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatal("Not a directory")
	}

	abs_meta := filepath.Join(abs_repo, "meta")

	info, err = os.Stat(abs_meta)

	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatal("Not a directory")
	}

	abs_data := filepath.Join(abs_repo, "data")

	info, err = os.Stat(abs_data)

	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatal("Not a directory")
	}

	throttle := make(chan bool, *limit)

	for i := 0; i < *limit; i++ {

		throttle <- true
	}

	var count int32
	count = 0

	var open int32
	open = 0

	callback := func(path string, info os.FileInfo) error {

		<-throttle

		defer func() {
			throttle <- true
		}()

		if info.IsDir() {
			return nil
		}

		alt, err := uri.IsAltFile(path)

		if err != nil {
			log.Fatal(err)
		}

		if alt {
			return nil
		}

		fh, err := os.Open(path)

		if err != nil {
			msg := fmt.Sprintf("Failed to open %s because %s (%d open filehandles)", path, err, atomic.LoadInt32(&open))
			log.Fatal(msg)
		}

		defer fh.Close()

		atomic.AddInt32(&open, 1)
		defer atomic.AddInt32(&open, -1)

		feature, err := ioutil.ReadAll(fh)

		if err != nil {
			log.Fatal(err)
		}

		meta.DumpFeature(feature)
		atomic.AddInt32(&count, 1)

		return nil
	}

	t1 := time.Now()

	cr := crawl.NewCrawler(abs_data)
	err = cr.Crawl(callback)

	t2 := time.Since(t1)
	log.Printf("time to dump %d features: %v\n", count, t2)

	if err != nil {
		log.Fatal(err)
	}
}
