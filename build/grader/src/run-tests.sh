#!/bin/sh
set -e
PART_ID=$1
if [ -z $PART_ID ]; then
  echo "PART_ID is empty"
  exit 1
fi

DIR="./$PART_ID"
if [ ! -d $DIR ]; then
  echo "Part $PART_ID is missing"
  exit 1
fi

cd $DIR
go clean -testcache
go test -cover -race -short ./...
echo "All tests passed"
