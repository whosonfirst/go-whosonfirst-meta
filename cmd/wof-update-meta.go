package main

import (
       "github.com/whosonfirst/go-whosonfirst-meta"
       "log"
       "os"
)

func main() {

     latest := "/usr/local/mapzen/whosonfirst-data/meta/wof-microhood-latest.csv"
     src, err := os.Open(latest)

     if err != nil {
     	log.Fatal(err)
     }
     
     dest := os.Stdout
     
     updated := make([]string, 0)

     meta.UpdateMetafile(src, dest, updated)
}
