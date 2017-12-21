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

dist-build:
	OS=darwin make dist-os
	OS=windows make dist-os
	OS=linux make dist-os

dist-os:
	mkdir -p dist/$(OS)
	GOOS=$(OS) GOPATH=$(GOPATH) GOARCH=386 go build -o dist/$(OS)/wof-build-metafiles cmd/wof-build-metafiles.go
	chmod +x dist/$(OS)/wof-build-metafiles
	cd dist/$(OS) && shasum -a 256 wof-build-metafiles > wof-build-metafiles.sha256

rmdist:
	if test -d dist; then rm -rf dist; fi

deps:   rmdeps
	@GOPATH=$(GOPATH) go get -u "github.com/facebookgo/atomicfile"
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/gjson"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-crawl"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-index"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-csv"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-placetypes"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-repo"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-uri"

vendor-deps: deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor/src; then rm -rf vendor/src; fi
	cp -r src vendor/src
	find vendor -name '.git' -print -type d -exec rm -rf {} +

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/wof-build-metafiles cmd/wof-build-metafiles.go
	@GOPATH=$(GOPATH) go build -o bin/wof-update-metafile cmd/wof-update-metafile.go
	@GOPATH=$(GOPATH) go build -o bin/wof-meta-prepare cmd/wof-meta-prepare.go

fmt:
	go fmt meta.go
	go fmt cmd/*.go
	go fmt meta/*.go
