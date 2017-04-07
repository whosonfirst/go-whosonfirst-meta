package main

import (
	"encoding/csv"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	var debug = flag.Bool("debug", false, "...")

	flag.Parse()

	writer := csv.NewWriter(os.Stdout)
	rows := 0

	for _, path := range flag.Args() {

		fh, err := os.Open(path)

		if err != nil {
			log.Fatal(err)
		}

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			log.Fatal(err)
		}

		row, err := meta.DumpFeature(body)

		if err != nil {
			log.Fatal(err)
		}

		defaults, err := meta.GetDefaults()

		if err != nil {
			log.Fatal(err)
		}

		row = defaults.EnsureDefaults(row)

		if *debug {

			for k, v := range row {

				log.Printf("[%s] '%s'\n", k, v)
			}

			continue
		}

		header := make([]string, 0)
		values := make([]string, 0)

		for k, v := range row {

			header = append(header, k)
			values = append(values, v)

			log.Println(k, v)
		}

		if rows == 0 {
			writer.Write(header)
		}

		writer.Write(values)
		writer.Flush()

		rows++
	}

}
