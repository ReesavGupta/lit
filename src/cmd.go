package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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

func HashObject() {

	/*
			git hash-object is used to compute the SHA hash of a Git object.
			When used with the -w flag, it also writes the object to the .git/objects directory.

			# Create a file with some content
		  $ echo -n "hello world" > test.txt

		  # Compute the SHA hash of the file + write it to .git/objects
		  $ git hash-object -w test.txt
		  95d09f2b10159347eece71399a7e2e907ea3df4f

		  # Verify that the file was written to .git/objects
		  $ file .git/objects/95/d09f2b10159347eece71399a7e2e907ea3df4f
		  .git/objects/95/d09f2b10159347eece71399a7e2e907ea3df4f: zlib compressed data
	*/

	fs := flag.NewFlagSet("hash-object", flag.ExitOnError)

	var write bool
	fs.BoolVar(&write, "w", false, "writes the hash object to the .git/objects dir")
	// lit hash-object -w text.txt
	fs.Parse(os.Args[2:])

	filePath := fs.Arg(0)

	content, err := os.ReadFile(filePath)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not open the file")
		os.Exit(1)
	}

	info, err := os.Stat(filePath)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not open the file")
		os.Exit(1)
	}

	headerAndContent := fmt.Sprintf("blob %d\x00%s", info.Size(), string(content))

	sha := sha1.Sum([]byte(headerAndContent))

	hash := fmt.Sprintf("%x", sha)

	blobName := []rune(hash)
	blobPath := "./.git/objects/"

	for i, v := range blobName {
		blobPath += string(v)
		if i == 1 {
			blobPath += "/"
		}
	}

	var buffer bytes.Buffer

	zw := zlib.NewWriter(&buffer)

	_, err = zw.Write([]byte(headerAndContent))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing compressing the header and content")
		zw.Close()
		os.Exit(1)
	}
	zw.Close()

	if write {
		err := os.MkdirAll(filepath.Dir(blobPath), os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create the directory ")
			os.Exit(1)
		}
		f, err := os.Create(blobPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating the file")
			os.Exit(1)
		}
		_, err = f.Write(buffer.Bytes())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing the contents of the buffer to the created file")
			os.Exit(1)
		}
		f.Close()
	}

	fmt.Print(hash)
}

func LsTree() {
	// this command is used to inspect a tree object
	// Trees are used to store directory structures.

	// A tree object has multiple entries:

	// ----> A SHA-1 hash that points to a blob or tree object
	// ---->---->If the entry is a file, this points to a blob object
	// ---->---->If the entry is a directory, this points to a tree object

	// ----> The name of the file/directory
	// ----> The mode of the file/directory
	// ---->----> This is a simplified version of the permissions you'd see in a Unix file system.
	// ---->----> For files, the valid values are:
	// ---->----> 100644 (regular file)
	// ---->----> 100755 (executable file)
	// ---->----> 120000 (symbolic link)
	// ---->----> For directories, the value is 40000
	// ---->----> There are other values for submodules, but we won't be dealing with those in this challenge

	// For exaple if you had a directory structure like:

	/*
		your_repo/
		- file1
		- dir1/
			- file_in_dir_1
			- file_in_dir_2
		- dir2/
			- file_in_dir_3
	*/

	// The entries in the tree object would look like this:

	// 40000 dir1 <tree_sha_1>
	// 40000 dir2 <tree_sha_2>
	// 100644 file1 <blob_sha_1>

	// Line 1 (40000 dir1 <tree_sha_1>) indicates that dir1 is a directory with the SHA hash <tree_sha_1>

	// Line 2 (40000 dir2 <tree_sha_2>) indicates that dir2 is a directory with the SHA hash <tree_sha_2>

	// Line 3 (100644 file1 <blob_sha_1>) indicates that file1 is a regular file with the SHA hash <blob_sha_1>

	// dir1 and dir2 would be tree objects themselves, and their entries would contain the files/directories inside them.

	/*
		Note that the output is alphabetically sorted, this is how Git stores entries in the tree object internally.
	*/

	// 	we'll implement the git ls-tree command with the --name-only flag. Here's how the output looks with the --name-only flag:

	//   $ git ls-tree --name-only <tree_sha>
	//   dir1
	//   dir2
	//   file1
	// The tester uses --name-only since this output format is easier to test against.

	/*
		Just like blobs, tree objects are stored in the .git/objects directory. If the hash of a tree object is e88f7a929cd70b0274c4ea33b209c97fa845fdbc, the path to the object would be

		./git/objects/e8/8f7a929cd70b0274c4ea33b209c97fa845fdbc
	*/

	// The format of a tree object file looks like this (after Zlib decompression):

	/*
		tree <size>\0
		<mode> <name>\0<20_byte_sha>
		<mode> <name>\0<20_byte_sha>
	*/

	// (The above code block is formatted with newlines for readability, but the actual file doesn't contain newlines)

	/*
		----> in a tree object file, the SHA-1 hashes are not in hexadecimal format. They're just raw bytes (20 bytes long).
		----> In a tree object file, entries are sorted by their name. The output of ls-tree matches this order.
	*/

	// $ lit ls-tree --name-only <tree_sha>

	var nameOnly bool
	fs := flag.NewFlagSet("name-only", flag.ExitOnError)

	fs.BoolVar(&nameOnly, "name-only", false, "outputs only the file names wihtout the header proper git object format ")

	fs.Parse(os.Args[2:])

	treeSha := fs.Arg(0)

	dir, fileName := treeSha[:2], treeSha[2:] //.git/object/dir/file

	filePath := fmt.Sprintf(".git/objects/%s/%s", dir, fileName)

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open object: %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	data, err := io.ReadAll(file)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
		os.Exit(1)
	}

	r, err := zlib.NewReader(bytes.NewReader(data))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failsed to decompress object: %v\n", err)
		os.Exit(1)
	}

	defer r.Close()

	decompressedData, err := io.ReadAll(r)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read decompressed data: %v\n", err)
		os.Exit(1)
	}

	// tree <size>\0<data>

	nullIndex := -1

	for i, v := range decompressedData {
		if v == 0 {
			nullIndex = i
			break
		}
	}

	if nullIndex == -1 {
		fmt.Fprintf(os.Stderr, "Invalid format")
		os.Exit(1)
	}

	treeContent := decompressedData[nullIndex+1:]

	// now we need to parse the entries
	// each entry is:

	// [mode] space [filename] null [20-bytes SHA]

	entries := []struct {
		Mode string
		Name string
		Sha  string
	}{}

	i := 0

	for i < len(treeContent) {
		// read mode
		modeStart := i
		for treeContent[i] != ' ' {
			i++
		}

		mode := string(treeContent[modeStart:i])

		i++ // skip space

		// Read filename
		nameStart := i
		for treeContent[i] != 0 {
			i++
		}
		name := string(treeContent[nameStart:i])
		i++ // skip null byte

		// Read SHA (20 bytes)
		shaBytes := treeContent[i : i+20]
		sha := fmt.Sprintf("%x", shaBytes)
		i += 20

		entries = append(entries, struct {
			Mode string
			Name string
			Sha  string
		}{mode, name, sha})

	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	for _, entry := range entries {
		if nameOnly {
			fmt.Println(entry.Name)
		} else {
			fmt.Printf("%s %s %s\n", entry.Mode, entry.Name, entry.Sha)
		}
	}

}
