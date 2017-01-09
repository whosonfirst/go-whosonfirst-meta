package main

import (
       "github.com/whosonfirst/go-whosonfirst-meta"
)

func main() {

     src := "/usr/local/mapzen/whosonfirst-data/meta/wof-microhood-latest.csv"
     dest := "foo.csv"
     
     updated := make([]string, 0)

     meta.UpdateMetafile(src, dest, updated)
}
