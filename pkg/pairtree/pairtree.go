/*
The Pairtree package will be utilized by both our command line and our
pairtree-service project
*/
package pairtree

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	caltech_pairtree "github.com/caltechlibrary/pairtree"
)

// File is the directory tree in JSON
type File struct {
	Name string `json:"name"`
}

// Directory is a directory file structure that can be nested
type Directory struct {
	Name        string      `json:"name"`
	Directories []Directory `json:"directories"`
	Files       []File      `json:"files"`
}

const (
	rootDir   = "pairtree_root"
	prefixDir = "pairtree_prefix"
	verDir    = "pairtree_version0_1"
)

// IsHidden determines if a file is hidden based on its name.
func IsHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

// IsDirectory determines if an object is a directory
func IsDirectory(obj fs.DirEntry) bool {
	return obj.IsDir()
}

// GetPrefix reads the content of the file at the pairtree prefix path and returns it as a string
func GetPrefix(ptRoot string) (string, error) {
	path := filepath.Join(ptRoot, prefixDir)

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist, return empty string and no error
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Check if the content is empty
	if len(content) == 0 {
		return "", error_msgs.Err1
	}

	// Return the content as a string
	return string(content), nil
}

// CheckPTVer checks if the pairtree_version0_1 is populated
func CheckPTVer(ptRoot string) error {
	path := filepath.Join(ptRoot, verDir)
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// Check if the file is empty
	if fileInfo.Size() == 0 {
		return error_msgs.Err2
	} else {
		return nil
	}
}

// CreatePP creates the full pairpath given the root, id, and prefix
func CreatePP(id, ptRoot, prefix string) (string, error) {
	if strings.TrimSpace(ptRoot) == "" {
		return "", error_msgs.Err3
	}

	if strings.TrimSpace(id) == "" {
		return "", error_msgs.Err4
	}

	if strings.HasPrefix(id, prefix) {
		// Remove the prefix from id
		id = strings.TrimPrefix(id, prefix)
	} else {
		return "", fmt.Errorf("%w, id: '%s', prefix: '%s'", error_msgs.Err5, id, prefix)
	}

	ptRoot = filepath.Join(ptRoot, rootDir)

	pairPath := caltech_pairtree.Encode(id)

	id = string(caltech_pairtree.CharEncode([]rune(id)))
	pairPath = filepath.Join(pairPath, id)
	pairPath = filepath.Join(ptRoot, pairPath)
	return pairPath, nil
}

// RecursiveFiles traverses directories recursively starting from the given pairPath and ID, returning a map
// where keys are directory paths and values are slices of fs.DirEntry. The traversal begins at the ID and
// recursively searches from that ID.
func RecursiveFiles(pairPath, id string) (map[string][]fs.DirEntry, error) {
	result := make(map[string][]fs.DirEntry)

	err := filepath.WalkDir(pairPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == pairPath {
			return nil
		}

		parentDir := filepath.Dir(path)

		// Add the directory entry to the map
		result[parentDir] = append(result[parentDir], d)

		// If the entry is a directory, initialize its entry in the map
		if d.IsDir() {
			result[path] = []fs.DirEntry{}
		}

		return nil
	})

	return result, err
}

// NonRecursiveFiles searches through a file structure non recursively
func NonRecursiveFiles(pairPath string) (map[string][]fs.DirEntry, error) {
	result := make(map[string][]fs.DirEntry)

	entries, err := os.ReadDir(pairPath)
	if err != nil {
		return nil, err
	}

	// Initialize the entry for the provided directory
	result[pairPath] = entries
	return result, nil
}

// BuildDirectoryTree recursively function to build the directory tree, isFirstIteration should always be
// set to true excpet for when it is being used recursively by BuildDirectoryTree()
func BuildDirectoryTree(path string, entriesMap map[string][]fs.DirEntry, isFirstIteration bool) Directory {
	var dir Directory
	path = filepath.FromSlash(path)
	if isFirstIteration {
		dir = Directory{
			Name: path, // Use the whole path name for the first iteration
		}
	} else {
		dir = Directory{
			Name: filepath.Base(path),
		}
	}

	for _, entry := range entriesMap[path] {
		if entry.IsDir() {
			subDirPath := filepath.Join(path, entry.Name())
			subDir := BuildDirectoryTree(subDirPath, entriesMap, false)
			dir.Directories = append(dir.Directories, subDir)
		} else {
			file := File{Name: entry.Name()}
			dir.Files = append(dir.Files, file)
		}
	}

	return dir
}

// ToJSONStructure converts the map into the desired JSON structure
func ToJSONStructure(dirTree Directory) ([]byte, error) {
	// Convert to JSON
	jsonData, err := json.MarshalIndent(dirTree, "", "  ")
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}
