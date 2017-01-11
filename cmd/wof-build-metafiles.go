package main

import (
	"flag"
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

	/*

		please to add a throttle because this:

		./bin/wof-build-metafiles -repo /usr/local/data/whosonfirst-data
		error: open /usr/local/data/whosonfirst-data/data/101/727/013: too many open files
		2017/01/11 00:16:58 time to dump 115213 features: 15.185776169s

	*/

	var count int32
	count = 0

	callback := func(path string, info os.FileInfo) error {

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
			log.Fatal(err)
		}

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
