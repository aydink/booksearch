#!/bin/bash

echo "$0-->$1" 

#echo "" > file_list.txt
cat /dev/null > file_list.txt
mkdir -p text
mkdir -p image

for filename in *.pdf 
do
    md5hash=$(md5 -q "$filename")
    echo "$md5hash $filename" >> file_list.txt
    mkdir -p image/$md5hash

    # convert to image
    #pdftocairo -jpeg "$filename" image/$md5hash/p
    pdftotext -enc UTF-8 "$filename" "text/$md5hash.txt" 
done
