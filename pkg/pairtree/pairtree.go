/*
The Pairtree package will be utilized by both our command line and our
pairtree-service project
*/
package pairtree

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	caltech_pairtree "github.com/caltechlibrary/pairtree"
	"github.com/mholt/archiver/v3"
	"github.com/otiai10/copy"
	"github.com/spf13/afero"
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
	PtPrefix  = "pt://"
	tar       = ".tgz"
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

// CreatePP creates the full pairpath given the root, id, and prefix giving the pairpath to an object
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

	// enocde ID to add to end of pairpath
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

// DeletePairtreeItem searches through a pairtree directory given the pairPath and subPath,
// and deletes the given directory or file.
func DeletePairtreeItem(fullPath string) error {
	// Check if the file or directory exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return err
	}

	// Attempt to remove the directory or file
	err := os.RemoveAll(fullPath)
	if err != nil {
		return err
	}
	return nil
}

// GetUniqueDestination checks if the destination path exists and appends ".x" (where x is an integer)
// to avoid overwriting files or directories.
func GetUniqueDestination(dest string) string {
	// If the destination does not exist, return it as is.
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return dest
	}

	// Extract the directory and base name
	dir := filepath.Dir(dest)
	base := filepath.Base(dest)

	// Strip the extension from the base name
	ext := filepath.Ext(base)
	baseWithoutExt := strings.TrimSuffix(base, ext)

	// Initialize counter for unique names
	counter := 1

	for {
		// Construct a new destination path by appending ".x" to the base name without extension
		newBase := fmt.Sprintf("%s.%d%s", baseWithoutExt, counter, ext)
		newDest := filepath.Join(dir, newBase)

		// If the new destination does not exist, return it
		if _, err := os.Stat(newDest); os.IsNotExist(err) {
			return newDest
		}
		counter++
	}
}

// CopyFileOrFolder copies a file or folder from src to dest, creating a unique destination if needed.
// It follows the same behavior as Unix cp with directories.
func CopyFileOrFolder(src, dest string, overwrite bool) (string, error) {
	// Get the source file or directory info
	_, err := os.Stat(src)
	if err != nil {
		return "", err
	}

	// If the destination is a directory, ensure it has the correct path
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		// If dest is a directory, append the base name of the source to dest
		dest = filepath.Join(dest, filepath.Base(src))
	} else if strings.HasSuffix(dest, string(os.PathSeparator)) {
		// If dest ends with '/', treat it as a directory
		dest = filepath.Join(dest, filepath.Base(src))
	}

	if !overwrite {
		// Ensure the destination path is unique
		dest = GetUniqueDestination(dest)
	}

	// Perform the copy operation using otiai10/copy
	err = copy.Copy(src, dest)
	if err != nil {
		return "", err
	}

	return dest, nil
}

// TarGz compresses the source directory or file into a .tgz archive.
// If the destination file already exists, it creates a unique destination.
// The prefix of the pairtree ID will be appended to the .tgz
func TarGz(src, dest, prefix string, overwrite bool) error {
	prefix = string(caltech_pairtree.CharEncode([]rune(prefix)))

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("could not create destination directory: %w", err)
	}

	dest = filepath.Join(dest, prefix+filepath.Base(src)+tar)

	if !overwrite {
		// Generate a unique destination if the file already exists
		dest = GetUniqueDestination(dest)
	}

	// Create a new archiver instance for tar.gz
	tgz := archiver.NewTarGz()

	// Archive the source directory
	if err := tgz.Archive([]string{src}, dest); err != nil {
		return fmt.Errorf("could not archive the source: %w", err)
	}

	return nil
}

// UnTarGz extracts a tar.gz archive to the specified destination directory.
// UntarGZ assumes that within the source .tgz file there is a folder that matches the name of
// the destination. If no such folder exists, UnTarGz will fail
func UnTarGz(src, dest string) error {
	id := filepath.Base(dest)
	fs := afero.NewOsFs()

	tempDir, err := afero.TempDir(fs, "", "temporary")
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, fs.RemoveAll(tempDir))
	}()

	// Create a TarGz archiver instance
	tgz := archiver.TarGz{
		Tar: &archiver.Tar{
			OverwriteExisting: true, // Keep this to handle file overwrites in case any remain
		},
	}

	// Extract the tar.gz archive to the destination directory
	if err := tgz.Unarchive(src, tempDir); err != nil {
		return err
	}

	// Check if tempDir contains a single folder that matches the pairtree ID
	files, err := afero.ReadDir(fs, tempDir)
	if err != nil {
		return fmt.Errorf("could not read temp directory: %w", err)
	}

	if len(files) != 1 || !files[0].IsDir() {
		return error_msgs.Err12
	}

	// Check if the folder name matches the pairtree ID
	if files[0].Name() != id {
		return error_msgs.Err13
	}

	// Ensure the source file exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return err
	}

	// Check if destination directory exists
	if _, err := os.Stat(dest); err == nil {
		// If it exists, clean up the destination directory to ensure full overwrite
		if err := os.RemoveAll(dest); err != nil {
			return err
		}
	}

	// Now you can move the folder from tempDir to the final destination
	if err := copy.Copy(filepath.Join(tempDir, id), dest); err != nil {
		return err
	}

	return err
}
