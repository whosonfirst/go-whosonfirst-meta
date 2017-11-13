#!/bin/sh

WHOAMI=`python -c 'import os, sys; print os.path.realpath(sys.argv[1])' $0`

if [ $? -ne 0 ]
then
    echo "[fatal] E_INSUFFICIENT_INTROSPECTION"
    exit 1
fi

ROOT=`dirname ${WHOAMI}`

GO=`which go`
CURL=`which curl`
UNZIP=`which unzip`

if [ ! -e ${GO} ]
then
    echo "You must have Go installed in order for this to work."
    exit 1
fi

cd ${ROOT}
export GOPATH=${ROOT}

for OS in darwin linux windows
do

    mkdir -p dist/${OS}
    rm -rf dist/${OS}/*

    # https://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5

    export GOOS=${OS}
    export GOARCH=386

    go build -o dist/${OS}/wof-build-metafiles cmd/wof-build-metafiles.go
done

exit 0
