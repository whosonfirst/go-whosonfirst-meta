CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test ! -d src; then mkdir src; fi
	if test ! -d src/github.com/whosonfirst/go-whosonfirst-meta; then mkdir -p src/github.com/whosonfirst/go-whosonfirst-meta/; fi
	cp  meta.go src/github.com/whosonfirst/go-whosonfirst-meta/
	cp -r meta src/github.com/whosonfirst/go-whosonfirst-meta/
	cp -r vendor/src/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:   rmdeps
	@GOPATH=$(GOPATH) go get -u "github.com/facebookgo/atomicfile"
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/gjson"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-crawl"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-csv"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-uri"

vendor-deps: deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor/src; then rm -rf vendor/src; fi
	cp -r src vendor/src
	find vendor -name '.git' -print -type d -exec rm -rf {} +

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/wof-build-metafiles cmd/wof-build-metafiles.go
	@GOPATH=$(GOPATH) go build -o bin/wof-update-metafile cmd/wof-update-metafile.go

fmt:
	go fmt meta.go
	go fmt cmd/*.go
