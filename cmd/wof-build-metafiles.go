package main

import (
	"flag"
	"fmt"
	"github.com/facebookgo/atomicfile"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"github.com/whosonfirst/go-whosonfirst-uri"
	_ "io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	procs := flag.Int("processes", runtime.NumCPU()*2, "The number of concurrent processes to use.")
	repo := flag.String("repo", "/usr/local/data/whosonfirst-data", "The repository to create metafiles for (and in).")
	limit := flag.Int("open-filehandles", 512, "The maximum number of file handles to keep open at any given moment.")

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

	// this is used below when creating file handles to write to
	// (20170410/thisisaaronland)

	repo_name := filepath.Base(abs_repo)
	repo_suffix := strings.Replace(repo_name, "whosonfirst-data-", "", -1)

	mu := new(sync.Mutex)

	throttle := make(chan bool, *limit)

	for i := 0; i < *limit; i++ {

		throttle <- true
	}

	var count int32
	var open int32
	var pending int32
	var scheduled int32

	count = 0
	open = 0
	pending = 0
	scheduled = 0

	filehandles := make(map[string]*atomicfile.File)
	writers := make(map[string]*csv.DictWriter)

	defer func() {

		for _, fh := range filehandles {
			fh.Close()
		}
	}()

	wg := new(sync.WaitGroup)

	callback := func(path string, info os.FileInfo) error {

		atomic.AddInt32(&pending, 1)
		// log.Printf("pending %d scheduled %d\n", pending, scheduled)

		<-throttle

		atomic.AddInt32(&pending, -1)
		atomic.AddInt32(&scheduled, 1)

		wg.Add(1)

		defer func() {
			throttle <- true
			wg.Done()
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

		row, err := meta.DumpFeature(feature)

		if err != nil {
			log.Fatal(err)
		}

		placetype := gjson.GetBytes(feature, "properties.wof:placetype").String()

		/*
		if placetype != "microhood" {
			return nil
		}
		*/

		mu.Lock()

		writer, ok := writers[placetype]

		if !ok {

			fieldnames := make([]string, 0)

			for k, _ := range row {
				fieldnames = append(fieldnames, k)
			}

			// repo_suffix is set above before we start processing
			// files (20170410/thisisaaronland)

			fname := fmt.Sprintf("wof-%s-latest.csv", placetype)

			if repo_suffix != "whosonfirst-data" {
				fname = fmt.Sprintf("wof-%s-latest.csv", repo_suffix)
			}

			outfile := filepath.Join("/tmp", fname)

			fh, err := atomicfile.New(outfile, os.FileMode(0644))

			if err != nil {
				log.Fatal(err)
			}

			writer, err = csv.NewDictWriter(fh, fieldnames)
			writer.WriteHeader()

			filehandles[placetype] = fh
			writers[placetype] = writer
		}

		writer.WriteRow(row)

		mu.Unlock()

		atomic.AddInt32(&count, 1)
		return nil
	}

	t1 := time.Now()

	cr := crawl.NewCrawler(abs_data)
	err = cr.Crawl(callback)

	wg.Wait()

	t2 := time.Since(t1)
	log.Printf("time to dump %d features: %v\n", count, t2)

	if err != nil {
		log.Fatal(err)
	}
}
