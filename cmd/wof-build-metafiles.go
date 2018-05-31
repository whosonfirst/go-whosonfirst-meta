package main

import (
       "context"
	"flag"
	"fmt"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"	
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"github.com/whosonfirst/go-whosonfirst-placetypes/filter"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	mode := flag.String("mode", "repo", "Where to read data (to create metafiles) from. If empty then the code will assume the current working directory.")
	out := flag.String("out", "", "Where to store metafiles. If empty then assume metafile are created in a child folder of 'repo' called 'meta'.")

	limit := flag.Int("open-filehandles", 512, "The maximum number of file handles to keep open at any given moment.")

	str_placetypes := flag.String("placetypes", "", "A comma-separated list of placetypes that meta files will be created for. All other placetypes will be ignored.")
	str_roles := flag.String("roles", "", "Role-based filters are not supported yet.")
	str_exclude := flag.String("exclude", "", "A comma-separated list of placetypes that meta files will not be created for.")

	timings := flag.Bool("timings", false, "...")

	procs := flag.Int("processes", 0, "The number of concurrent processes to use. THIS FLAG HAS BEEN DEPRECATED")

	flag.Parse()

	placetypes := make([]string, 0)
	roles := make([]string, 0)
	exclude := make([]string, 0)

	if *str_placetypes != "" {
		placetypes = strings.Split(*str_placetypes, ",")
	}

	if *str_roles != "" {
		roles = strings.Split(*str_roles, ",")
	}

	if *str_exclude != "" {
		exclude = strings.Split(*str_exclude, ",")
	}

	placetype_filter, err := filter.NewPlacetypesFilter(placetypes, roles, exclude)

	if err != nil {
		log.Fatal(err)
	}

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

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

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

		path, err := index.PathForContext(ctx)

		if err != nil {
			return err
		}

		ok, err := utils.IsPrincipalWOFRecord(fh, ctx)

		if err != nil {
			return err
		}

		if !ok {
			return nil
		}

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil {
			return err
		}

		atomic.AddInt32(&open, 1)
		defer atomic.AddInt32(&open, -1)

		feature, err := ioutil.ReadAll(fh)

		if err != nil {
			log.Fatal(err)
		}

		placetype := f.Placetype()

		allow, err := placetype_filter.AllowFromString(placetype)

		if err != nil {
			log.Println(fmt.Sprintf("Unable to validate placetype (%s) for %s", placetype, path))
			return err
		}

		if !allow {
			return nil
		}

		row, err := meta.FeatureToRow(feature)

		if err != nil {
			return err
		}

		r, err := repo.NewDataRepoFromString(whosonfirst.Repo(f))

		if err != nil {
			return err		   
		}
		
		mu.Lock()

		writer, ok := writers[placetype]

		if !ok {

			fieldnames := make([]string, 0)

			for k, _ := range row {
				fieldnames = append(fieldnames, k)
			}

			sort.Strings(fieldnames)

			opts := repo.DefaultFilenameOptions()
			opts.Placetype = placetype

			fname := r.MetaFilename(opts)

			outfile := filepath.Join(abs_meta, fname)

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

	i, err := index.NewIndexer(*mode, cb)

	if err != nil {
		log.Fatal(err)
	}

	t1 := time.Now()

	for _, path := range flag.Args() {

		ta := time.Now()

		err := i.IndexPath(path)

		if err != nil {
			log.Fatal(err)
		}

		tb := time.Since(ta)

		if *timings {
			log.Printf("time to prepare %s %v\n", path, tb)
		}

	}

	t2 := time.Since(t1)

	if *timings {
		c := atomic.LoadInt32(&count)
		log.Printf("time to prepare all %d records %v\n", c, t2)
	}
}
