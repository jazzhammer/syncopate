package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func emptyDir(dir string) {
	items, _ := os.ReadDir(dir)
	for _, item := range items {
		os.RemoveAll(fmt.Sprintf("%s%c%s", dir, filepath.Separator, item.Name()))
	}
}

func copySource(src string, dest string) {
	srcBytes, _ := os.ReadFile(src)
	os.WriteFile(dest, srcBytes, 0644)
}

// Command given: sync -s source_folder_a -d target_folder
// Expected result: target_folder will have the exact same contents with source_folder_a
func AtoEmptyTarget(t *testing.T) {
	emptyDir("source_folder_a")
	copySource("all_sources/file_a", "source_folder_a/file_a")
	copySource("all_sources/file_b", "source_folder_a/file_b")
	copySource("all_sources/file_c", "source_folder_a/file_c")

	emptyDir("target_folder")
	os.Args = []string{"sync", "-s", "source_folder_a", "-d", "target_folder"}
	main()
	compareSourceDestinationFolders("source_folder_a", "target_folder", t, false)
}

func compareSourceDestinationFolders(src string, dest string, t *testing.T, filesOnly bool) {
	itemsSrc, _ := os.ReadDir(src)
	itemsDest, _ := os.ReadDir(dest)
	if filesOnly {
		var filteredSrc = []os.DirEntry{}
		for _, itemSrc := range itemsSrc {
			if !itemSrc.IsDir() {
				filteredSrc = append(filteredSrc, itemSrc)
			}
		}
		itemsSrc = filteredSrc
		var filteredDest = []os.DirEntry{}
		for _, itemDest := range itemsDest {
			if !itemDest.IsDir() {
				filteredDest = append(filteredDest, itemDest)
			}
		}
		itemsDest = filteredDest
	}
	if len(itemsSrc) != len(itemsDest) {
		t.Errorf("copy produced %d items. expected %d", len(itemsDest), len(itemsSrc))
	}
	for i, itemA := range itemsSrc {
		if itemA.Name() != itemsDest[i].Name() {
			t.Errorf("copy produced unexpected file: %c%s", filepath.Separator, itemsDest[i].Name())
			break
		}
	}
	// recurse to inspect folders
	for s, itemSrc := range itemsSrc {
		if itemSrc.IsDir() {
			compareSourceDestinationFolders(
				fmt.Sprintf("%s%c%s", src, filepath.Separator, itemSrc.Name()),
				fmt.Sprintf("%s%c%s", dest, filepath.Separator, itemsDest[s].Name()),
				t,
				false,
			)
		}
	}
}

// Command given: sync -s source_folder_b -d target_folder
// Expected result: target_folder will have the exact same contents with source_folder_b, meaning
//
//	file_b is deleted and
//	file_d is copied.
func BtoAPresynced(t *testing.T) {
	emptyDir("source_folder_b")
	copySource("all_sources/file_a", "source_folder_b/file_a")
	copySource("all_sources/file_c", "source_folder_b/file_c")
	copySource("all_sources/file_d", "source_folder_b/file_d")

	emptyDir("source_folder_a")
	copySource("all_sources/file_a", "source_folder_a/file_a")
	copySource("all_sources/file_b", "source_folder_a/file_b")
	copySource("all_sources/file_c", "source_folder_a/file_c")

	os.Args = []string{"sync", "-s", "source_folder_b", "-d", "source_folder_a"}
	main()
	compareSourceDestinationFolders("source_folder_b", "source_folder_a", t, false)
}

// Command given: sync -s source_folder_c -d target_folder
// Expected result: target_folder will have the exact same contents with source_folder_c, meaning:
//
//	file_c is deleted,
//	file_e is copied into root and
//	file_a_a is copied into dir_a.
func CtoBPresynced(t *testing.T) {
	emptyDir("source_folder_c")
	copySource("all_sources/file_a", "source_folder_c/file_a")
	copySource("all_sources/file_d", "source_folder_c/file_d")
	copySource("all_sources/file_e", "source_folder_c/file_e")
	os.MkdirAll("source_folder_c/dir_a", 0700)
	copySource("all_sources/file_a_a", "source_folder_c/dir_a/file_a_a")

	emptyDir("source_folder_b")
	copySource("all_sources/file_a", "source_folder_b/file_a")
	copySource("all_sources/file_c", "source_folder_b/file_c")
	copySource("all_sources/file_d", "source_folder_b/file_d")

	os.Args = []string{"sync", "-s", "source_folder_c", "-d", "source_folder_b"}
	main()
	compareSourceDestinationFolders("source_folder_c", "source_folder_b", t, false)
}

// Command given: sync -s source_folder_a -d target_folder
// Expected result: target_folder will have the exact same contents with source_folder_a, meaning:
// file_d, file_e and file_a_a are deleted and
// file_b and file_c are copied.
func AtoCPresynced(t *testing.T) {
	emptyDir("source_folder_a")
	copySource("all_sources/file_a", "source_folder_a/file_a")
	copySource("all_sources/file_b", "source_folder_a/file_b")
	copySource("all_sources/file_c", "source_folder_a/file_c")

	emptyDir("source_folder_c")
	copySource("all_sources/file_a", "source_folder_c/file_a")
	copySource("all_sources/file_d", "source_folder_c/file_d")
	copySource("all_sources/file_e", "source_folder_c/file_e")
	os.MkdirAll("source_folder_c/dir_a", 0700)
	copySource("all_sources/file_a_a", "source_folder_c/dir_a/file_a_a")

	os.Args = []string{"sync", "-s", "source_folder_a", "-d", "source_folder_c"}
	main()
	compareSourceDestinationFolders("source_folder_a", "source_folder_c", t, false)
}

// constraint
// Consider only files, empty directories don't matter.
// nb: calls compareSourceDestinationFolders with filesOnly = true, as the empty folder should not be copied
func IgnoreEmptyFolder(t *testing.T) {
	emptyDir("source_folder_c")
	copySource("all_sources/file_a", "source_folder_c/file_a")
	copySource("all_sources/file_d", "source_folder_c/file_d")
	copySource("all_sources/file_e", "source_folder_c/file_e")
	os.MkdirAll("source_folder_c/dir_a", 0700)

	emptyDir("source_folder_a")
	copySource("all_sources/file_a", "source_folder_a/file_a")
	copySource("all_sources/file_d", "source_folder_a/file_d")
	copySource("all_sources/file_e", "source_folder_a/file_e")
	os.MkdirAll("source_folder_a/dir_a", 0700)
	copySource("all_sources/file_a_a", "source_folder_a/dir_a/file_a_a")

	os.Args = []string{"sync", "-s", "source_folder_c", "-d", "source_folder_a"}
	main()
	compareSourceDestinationFolders("source_folder_c", "source_folder_a", t, true)
}

func TestSync(t *testing.T) {
	AtoEmptyTarget(t)
	BtoAPresynced(t)
	CtoBPresynced(t)
	AtoCPresynced(t)
	IgnoreEmptyFolder(t)
}
