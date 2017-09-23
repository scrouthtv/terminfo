#!/bin/bash

set -e

go build

SI=$(which infocmp)
BI=./infocmp

OUT=.out

mkdir -p $OUT

TERMS=$(find /usr/share/terminfo -type f)
for i in $TERMS; do
  NAME=$(basename $i)

  TERM=$NAME $SI -1 -L -x > $OUT/$NAME-orig.txt
  TERM=$NAME $BI -x > $OUT/$NAME-test.txt
done

md5sum $OUT/*-orig.txt > $OUT/orig.md5

cp $OUT/orig.md5 $OUT/test.md5

perl -pi -e 's/-orig\.txt$/-test.txt/g' $OUT/test.md5

for i in $(md5sum --quiet -c $OUT/test.md5 |grep FAILED|sed -s 's/: FAILED//'); do
  $HOME/src/misc/icdiff/icdiff $(sed -e 's/-test\.txt$/-orig.txt/' <<< "$i") $i
done
