package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
)

// getDotFilePath returns the dot file for the repos list.
// Creates it add the enclosing folder if it does not exists
func getDotFilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dotFile := usr.HomeDir + "/.gogitlocalstats"

	return dotFile
}

// openFile opens the file located at `filePath`. Creates ii if not existing
func openFileReadOnly(filePath string) *os.File {
	f, err := os.Open(filePath) // O_RDONLY
	if err != nil {
		if os.IsNotExist(err) {
			// если файла нет — создаём пустой и открываем заново для чтения
			if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
				panic(err)
			}
			f, err = os.Open(filePath)
		}
	}

	if err != nil {
		panic(err)
	}

	return f
}

// parseFileLinesToSlice given a file path string, gets the content
// of each line and parses it to a slice of string
func parseFileLinesToSlice(filePath string) []string {
	f := openFileReadOnly(filePath)
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	return lines
}

// sliceContains returns true if `slice` contaings `value`
func sliceContains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}

// joinSlices adds the element of the `new` slice
// into the `existing` slice, only if not already there
func joinSlices(new []string, existing []string) []string {
	for _, i := range new {
		if !sliceContains(existing, i) {
			existing = append(existing, i)
		}
	}

	return existing
}

// dumpStringsSliceToFile writes content to the file in path `filePath` (overwriting existing content)
func dumpStringSliceToFile(repos []string, filePath string) {
	content := strings.Join(repos, "\n")
	os.WriteFile(filePath, []byte(content), 0755)
}

// addNewSliceElemetsToFile given a slice of strings representing paths, stores them
// to the filesystem
func addNewSliceElementsToFile(filePath string, newRepos []string) {
	existingRepos := parseFileLinesToSlice(filePath)
	repos := joinSlices(newRepos, existingRepos)
	dumpStringSliceToFile(repos, filePath)
}

// recursiveScanFoldeer starts the recursive search of git repositories
// living in the `folder` subtree
func recursiveScanFolder(folder string) []string {
	return scanGitFolders(make([]string, 0), folder)
}

// scanGitFolders returns a list of subfolders of `folder` ending with `.git`.
// Returns the base folder of the repo, the .git folder parent.
// Recursively searches in the subfolders by passing an existing `folders` slice.
func scanGitFolders(folders []string, folder string) []string {
	// trim the last `/`
	folder = strings.TrimSuffix(folder, "/")

	f, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	files, err := f.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}

	var path string

	for _, file := range files {
		if file.IsDir() {
			path = folder + "/" + file.Name()
			if file.Name() == ".git" {
				path = strings.TrimSuffix(path, "/.git")
				fmt.Println(path)
				folders = append(folders, path)
				continue
			}

			if file.Name() == "vendor" || file.Name() == "node_modules" {
				continue
			}

			folders = scanGitFolders(folders, path)
		}
	}

	return folders
}

// scan scans a new folder for Git repositories
func scan(folder string) {
	fmt.Printf("Found folders:\n\n")
	repositories := recursiveScanFolder(folder)
	filePath := getDotFilePath()
	addNewSliceElementsToFile(filePath, repositories)
	fmt.Printf("\n\nSuccessfully added\n\n")
}
