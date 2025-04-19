package main

import (
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"os"
)

func InitRepository() {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
	}

	fmt.Println("Initialized git directory")
}

func CatFile() {
	// This function should retrieve the content of the file from the git object store.
	// checks if the pretty-print flag is present or not.
	// if -p provided we print the content else we print the header

	var prettyPrint bool // flag variable

	// the default flag parser (flag.Parse()) only works once per program, and always parses os.Args[1:]
	// create a new flagset
	fs := flag.NewFlagSet("cat-file", flag.ExitOnError)
	fs.BoolVar(&prettyPrint, "p", false, "pretty-print")
	fs.Parse(os.Args[2:]) //now will parse for flags only after the second argument

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: mygit cat-file -flag<optional> <SHA>")
		os.Exit(1)
	}

	fileSHA := fs.Arg(0)

	// the object store holds the sha1 hash blob inside directories of their initial two chars. the file name is the rest of the hash
	dir, fileName := fileSHA[:2], fileSHA[2:]

	filePath := fmt.Sprintf(".git/objects/%s/%s", dir, fileName)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open object: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader, err := zlib.NewReader(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create zlib reader: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read object content: %v\n", err)
		os.Exit(1)
	}

	// The data format is: "blob <size>\0<content>"
	nullIndex := -1

	for i, b := range content {
		if b == 0 {
			nullIndex = i
			break
		}
	}

	if nullIndex == -1 {
		fmt.Fprintln(os.Stderr, "Invalid Git object format (no null byte found)")
		os.Exit(1)
	}
	header := string(content[:nullIndex]) // e.g., "blob 16"
	body := string(content[nullIndex+1:]) // actual file content
	if prettyPrint {
		fmt.Print(body)
	} else {
		fmt.Print(header)
	}

}
