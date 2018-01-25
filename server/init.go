package main

import (
	"bufio"
	"log"
	"os"
)

// fileMap will hold a mapping between md5 of a file and its full path
var fileMap map[string]string

func loadFileList() map[string]string {

	fileMap := make(map[string]string)

	file, err := os.Open("../scripts/file_list.txt")
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		md5hash := line[0:32]
		filename := line[33:]

		fileMap[md5hash] = filename
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return fileMap
}

// before the senver accept connection load file mapping from disk
func init() {
	fileMap = loadFileList()
}
