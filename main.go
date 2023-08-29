package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var flag_verbose = false

/*
dumps banner and help text to console
*/
func outputHelp() {
	if !flag_verbose {
		return
	}
	fmt.Println(`============================================================
  _________.__. ____   ____
 /  ___<   |  |/    \_/ ___\
 \___ \ \___  |   |  \  \___
/____  >/ ____|___|  /\___  >
     \/ \/         \/     \/
------------------------------------------------------------
sync utility

copy files from a source directory to a destination directory

usage:

> sync [options]

-s  required    source directory
-d  required    destination directory`)
}

/*
dir: directory to be inspected
returns:

	slice of strings = directories in dir, minus empty directories
	slice of strings = files in dir
*/
func getDirEntrys(dir string) ([]string, []string) {
	entrys, _ := os.ReadDir(dir)
	folders := []string{}
	files := []string{}
	for _, entry := range entrys {
		path := fmt.Sprintf("%s%c%s", dir, filepath.Separator, entry.Name())
		if entry.IsDir() {
			isDir := fmt.Sprintf("%s%c%s", dir, filepath.Separator, entry.Name())
			dirEntrys, err := os.ReadDir(isDir)
			if err != nil {
				error(fmt.Sprintf("error while inspecting folder: %s", isDir))
			}
			if len(dirEntrys) > 0 {
				folders = append(folders, path)
			}
			folders = append(folders, path)
		} else {
			files = append(files, path)
		}
	}
	if flag_verbose {
		if len(folders) > 0 || len(files) > 0 {
			fmt.Printf("d, f = %d, %d, dir='%s'\n", len(folders), len(files), dir)
		}
	}
	return folders, files
}

/*
validates source and destination as directories
returns:

	validated source folder string,
	destination folder string

or

	exits with code 2 for empty or invalid source or destination
*/
func parseArgs() (source string, destination string) {
	args := os.Args
	var dest string
	var src string
	var parsing = "any"
	for _, arg := range args {
		if strings.HasPrefix(arg, "-v") {
			flag_verbose = true
			continue
		}
		if parsing == "destination" {
			if !strings.HasPrefix(arg, "-") {
				dest = arg
			}
			parsing = "any"
		} else if parsing == "source" {
			if !strings.HasPrefix(arg, "-") {
				src = arg
			}
			parsing = "any"
		}
		if strings.HasPrefix(strings.ToLower(arg), "-d") {
			if arg == "-d" {
				parsing = "destination"
			} else {
				if dest == "" {
					dest = arg[2:]
				} else {
					status(fmt.Sprintf("repeated param -d ignored: %s\n", arg[2:]))
				}
			}
		}
		if strings.HasPrefix(strings.ToLower(arg), "-s") {
			if arg == "-s" {
				parsing = "source"
			} else {
				if src == "" {
					src = arg[2:]
				} else {
					status(fmt.Sprintf("repeated param -s ignored: %s\n", arg[2:]))
				}
			}
		}
	}
	if src == "" || dest == "" {
		if src == "" {
			error("option -s <source folder> required")
		}
		if dest == "" {
			error("option -d <destination folder> required")
		}
		outputHelp()
		os.Exit(2)
	}
	reader, _ := os.Open(src)
	_, err := reader.Readdir(0)
	if err != nil {
		error(fmt.Sprintf("-s(source) is not a directory: %s\n", src))
		os.Exit(2)
	}
	reader, _ = os.Open(dest)
	_, err = reader.Readdir(0)
	if err != nil {
		error(fmt.Sprintf("-d(destination) is not a directory: %s\n", dest))
		os.Exit(2)
	}
	return src, dest
}

func copyFiles(paths []string, src string, dest string) {
	if len(paths) > 0 {
		status(fmt.Sprintf("copy files: %d:", len(paths)))
		for _, path := range paths {
			// avoid copy if files are identical
			// the src file to compare
			fileInfoSrc, err := os.Stat(path)
			if err != nil {
				error(fmt.Sprintf("failed to inspect source file: %s, %s", path, err.Error()))
				os.Exit(2)
			}
			// the dest file to compare
			destPath := fmt.Sprintf("%s%s", dest, path[len(src):])
			destFolder := destPath[:strings.LastIndexAny(destPath, string(filepath.Separator))]
			if _, err := os.Stat(destFolder); err != nil {
				os.MkdirAll(destFolder, 0700)
			}
			fileInfoDest, err := os.Stat(destPath)

			if os.SameFile(fileInfoSrc, fileInfoDest) {
				status(fmt.Sprintf("skip identical: %s->%s", path, destPath))
				continue
			} else {
				copyFile(path, destPath)
			}
		}
	}
}

func deleteUnmatchedFiles(paths []string, src string, dest string) {
	if len(paths) > 0 {
		// hash files in source folder
		amongSource := make(map[string]string)
		for _, path := range paths {
			filename := path[(strings.LastIndexAny(path, string(filepath.Separator)) + 1):]
			amongSource[filename] = path
		}

		// files in dest folder
		destPath := fmt.Sprintf("%s%s", dest, paths[0][len(src):])
		destFolder := destPath[:strings.LastIndexAny(destPath, string(filepath.Separator))]
		if _, err := os.Stat(destFolder); err != nil {
			os.MkdirAll(destFolder, 0700)
		}
		destEntrys, err := os.ReadDir(destFolder)
		if err != nil {
			error(fmt.Sprintf("error reading target folder: %s", err.Error()))
			os.Exit(2)
		}

		for _, destEntry := range destEntrys {
			_, among := amongSource[destEntry.Name()]
			if !among {
				err := os.RemoveAll(fmt.Sprintf("%s%c%s", destFolder, filepath.Separator, destEntry.Name()))
				if err != nil {
					error(fmt.Sprintf("error deleting target file: %s", err.Error()))
					os.Exit(2)
				}
			}
		}
	}
}

func copyFile(source string, target string) {
	status(fmt.Sprintf("copy: %s -> %s", source, target))

	bytes := make([]byte, 2048)

	file, err := os.Open(source)
	if err != nil {
		error(err.Error())
		file.Close()
		os.Exit(2)
	}

	destFile, err := os.Create(target)
	for {
		red, err := file.Read(bytes)
		if err != nil {
			if err.Error() != "EOF" {
				error(err.Error())
				file.Close()
				os.Exit(2)
			}
		}
		if red > 0 {
			_, err := destFile.Write(bytes[:red])
			if err != nil {
				error(fmt.Sprintf("%s writing %s", err.Error(), target))
				destFile.Close()
				file.Close()
				os.Exit(2)
			}
		} else {
			break
		}
	}
	destFile.Close()
	file.Close()
}

func status(output string) {
	if flag_verbose {
		fmt.Println(output)
	}
}

func error(output string) {
	fmt.Println(output)
}

func main() {
	src, dst := parseArgs()
	status(fmt.Sprintf("--------------------------------------------------"))
	status(fmt.Sprintf("sync 0.1 %s -> %s\n", src, dst))
	status(fmt.Sprintf("--------------------------------------------------"))
	folders, files := getDirEntrys(src)
	for len(folders) > 0 || len(files) > 0 {
		deleteUnmatchedFiles(files, src, dst)
		copyFiles(files, src, dst)
		files = []string{}
		if len(folders) > 0 {
			folder := folders[0]
			status(fmt.Sprintf("folders: %d, inspect %s\n", len(folders), folder))
			folders = folders[1:]
			nextFolders, nextFiles := getDirEntrys(folder)
			folders = append(folders, nextFolders...)
			files = nextFiles
		}
	}
}
