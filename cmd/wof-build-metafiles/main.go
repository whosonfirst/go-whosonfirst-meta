package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-meta/build"
	"github.com/whosonfirst/go-whosonfirst-meta/options"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	valid_modes := strings.Join(wof_index.Modes(), ",")
	desc_modes := fmt.Sprintf("The mode to use importing data. Valid modes are: %s.", valid_modes)
	
	mode := flag.String("mode", "repo", desc_modes)
	out := flag.String("out", "", "Where to store metafiles. If empty then assume metafile are created in the current working directory.")

	// blah blah blah interface definition mismatch atomicfile blah blah blah...
	// stdout := flag.Bool("stdout", false, "Write meta file(s) to STDOUT")

	limit := flag.Int("open-filehandles", 512, "The maximum number of file handles to keep open at any given moment.")

	str_placetypes := flag.String("placetypes", "", "A comma-separated list of placetypes that meta files will be created for. All other placetypes will be ignored.")
	str_roles := flag.String("roles", "", "Role-based filters are not supported yet.")
	str_exclude := flag.String("exclude", "", "A comma-separated list of placetypes that meta files will not be created for.")

	timings := flag.Bool("timings", false, "Display timings during processing.")
	strict := flag.Bool("strict", false, "Be strict about requiring defined WOF placetypes.")

	old_skool := flag.Bool("old-skool", false, "Use old-skool wof-{PLACETYPE}-latest filenames")

	procs := flag.Int("processes", 0, "The number of concurrent processes to use. THIS FLAG HAS BEEN DEPRECATED")

	flag.Parse()

	if *procs != 0 {
		log.Println("the -procs flag has been deprecated and will be ignored")
	}

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

	var abs_root string

	if *out == "" {

		cwd, err := os.Getwd()

		if err != nil {
			log.Fatal(err)
		}

		abs_root = cwd
	} else {

		abs_out, err := filepath.Abs(*out)

		if err != nil {
			log.Fatal(err)
		}

		abs_root = abs_out
	}

	opts, err := options.DefaultBuildOptions()

	if err != nil {
		log.Fatal(err)
	}

	opts.Placetypes = placetypes
	opts.Roles = roles
	opts.Exclude = exclude
	opts.Workdir = abs_root
	opts.Timings = *timings
	opts.Strict = *strict
	opts.MaxFilehandles = *limit
	opts.OldSkool = *old_skool

	paths := flag.Args()

	metafiles, err := build.BuildFromIndex(opts, *mode, paths)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(metafiles)
}
