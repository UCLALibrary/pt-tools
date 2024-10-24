package pairtree

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	"github.com/UCLALibrary/pt-tools/testutils"
	"github.com/mholt/archiver/v3"
	"github.com/otiai10/copy"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	prefix = "ark:/"
)

// Dummy implementation of fs.DirEntry for testing purposes
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return m.isDir }
func (m mockDirEntry) Type() fs.FileMode          { return 0 }
func (m mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// updateMapKeys adds a prefix to all keys in the map.
func updateMapKeys(original map[string][]fs.DirEntry, prefix string) map[string][]fs.DirEntry {
	newMap := make(map[string][]fs.DirEntry)
	for k, v := range original {
		newKey := filepath.Join(prefix, k)
		newMap[newKey] = v
	}
	return newMap
}

// DeleteFileInDir deletes the specified file from the given directory.
func DeleteFileInDir(dir string, filename string) error {
	// Join the directory and filename to get the full path of the file to delete.
	filePath := filepath.Join(dir, filename)

	// Delete the file.
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// CompareDirEntries compares two fs.DirEntry instances.
func CompareDirEntries(a, b fs.DirEntry) bool {
	return a.Name() == b.Name() && a.IsDir() == b.IsDir()
}

// CompareDirEntrySlices compares two slices of fs.DirEntry, treating them as sets.
func CompareDirEntrySlices(slice1, slice2 []fs.DirEntry) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	// Convert slices to maps for easier comparison
	entryMap1 := make(map[string]fs.DirEntry)
	entryMap2 := make(map[string]fs.DirEntry)

	for _, entry := range slice1 {
		entryMap1[entry.Name()] = entry
	}

	for _, entry := range slice2 {
		entryMap2[entry.Name()] = entry
	}

	if len(entryMap1) != len(entryMap2) {
		return false
	}

	for name, entry1 := range entryMap1 {
		entry2, exists := entryMap2[name]
		if !exists || !CompareDirEntries(entry1, entry2) {
			return false
		}
	}

	return true
}

// CompareMaps compares two maps with string keys and []fs.DirEntry values.
func CompareMaps(map1, map2 map[string][]fs.DirEntry) bool {
	if len(map1) != len(map2) {
		return false
	}

	for key, entries1 := range map1 {
		entries2, exists := map2[key]
		if !exists || !CompareDirEntrySlices(entries1, entries2) {
			return false
		}
	}

	return true
}

// compareDirectories compares two Directory structs for equality
func compareDirectories(a, b Directory) bool {
	if a.Name != b.Name {
		return false
	}
	if len(a.Directories) != len(b.Directories) {
		return false
	}
	if len(a.Files) != len(b.Files) {
		return false
	}
	for i := range a.Directories {
		if !compareDirectories(a.Directories[i], b.Directories[i]) {
			return false
		}
	}
	for i := range a.Files {
		if a.Files[i].Name != b.Files[i].Name {
			return false
		}
	}
	return true
}

// TestIsHidden tests the IsHidden() function
func TestIsHidden(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{name: ".hiddenfile.txt", expected: true},
		{name: ".hiddenfolder", expected: true},
		{name: "visiblefile", expected: false},
		{name: ".hiddenfolder", expected: true},
		{name: "subdir/.hidden.txt", expected: false},
		{name: "", expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsHidden(test.name)
			assert.Equal(t, test.expected, result)
		})
	}
}

// TestGetPrefix creates a temporary directory with Afero and alters the prefix file depending on test needs
func TestGetPrefix(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		expectPre   string
		expectError error
	}{
		{
			name:        "noPrefix",
			expectPre:   "",
			expectError: error_msgs.Err1,
		},
		{
			name:        "prefixExists",
			expectPre:   prefix,
			expectError: nil,
		},
		{
			name:        "noPrefixFile",
			expectPre:   "",
			expectError: nil,
		},
	}

	fs := afero.NewOsFs()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a temporary directory for this test
			tempDir := testutils.CreateTempDir(t, fs)

			// Copies entire directory in testutils.TestPairtree into the temporary directory
			testutils.CopyTestDirectory(t, testutils.TestPairtree, tempDir)

			prefixFile := filepath.Join(tempDir, prefixDir)

			if test.name == "noPrefixFile" {
				err := fs.Remove(prefixFile)
				if err != nil {
					t.Logf("Error deleting file %s: %v", prefixFile, err)
				}
			} else if test.name == "noPrefix" {
				err := afero.WriteFile(fs, prefixFile, []byte{}, 0644)
				if err != nil {
					t.Logf("Error clearing file %s: %v", prefixFile, err)
				}
			}

			pre, err := GetPrefix(tempDir)
			assert.Equal(t, test.expectPre, pre)
			assert.ErrorIs(t, err, test.expectError)
		})
	}
}

// TestCreatePP tests various senarios related to creating a pairpath
func TestCreatePP(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		ptRoot    string
		prefix    string
		expectErr error
		expectPP  []string
	}{
		{
			name:      "standard",
			id:        "ark:/345621",
			ptRoot:    "root",
			prefix:    prefix,
			expectErr: nil,
			expectPP:  []string{"root", "pairtree_root", "34", "56", "21", "345621"},
		},
		{
			name:      "specialChars",
			id:        "ark:/34:621",
			ptRoot:    "root",
			prefix:    prefix,
			expectErr: nil,
			expectPP:  []string{"root", "pairtree_root", "34", "+6", "21", "34+621"},
		},
		{
			name:      "noPrefix",
			id:        "34621",
			ptRoot:    "root",
			prefix:    "",
			expectErr: nil,
			expectPP:  []string{"root", "pairtree_root", "34", "62", "1", "34621"},
		},
		{
			name:      "noPtRoot",
			id:        "34621",
			ptRoot:    "",
			prefix:    prefix,
			expectErr: error_msgs.Err3,
			expectPP:  nil,
		},
		{
			name:      "noId",
			id:        "",
			ptRoot:    "root",
			prefix:    "",
			expectErr: error_msgs.Err4,
			expectPP:  nil,
		},
		{
			name:      "idNoPrefix",
			id:        "34621",
			ptRoot:    "root",
			prefix:    prefix,
			expectErr: error_msgs.Err5,
			expectPP:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pairpath, err := CreatePP(test.id, test.ptRoot, test.prefix)

			// Join the expected path components
			expectedPairpath := ""
			if test.expectPP != nil {
				expectedPairpath = filepath.Join(test.expectPP...)
			}

			assert.Equal(t, expectedPairpath, pairpath)
			assert.ErrorIs(t, err, test.expectErr)

		})
	}
}

// TestGetPrefix creates a temporary directory with Afero and alters the prefix file depending on test needs
func TestRecursiveFiles(t *testing.T) {
	// Define test cases
	tests := []struct {
		pairpath    string
		id          string
		expectError error
		expectMap   map[string][]fs.DirEntry
	}{
		{
			pairpath:    filepath.Join("a5", "38", "8", "a5388"),
			id:          "a5388",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				filepath.Join("a5", "38", "8", "a5388"): {
					mockDirEntry{name: "a5388.txt", isDir: false},
				},
			},
		},
		{
			pairpath:    filepath.Join("a5", "48", "92", "a54892"),
			id:          "a54892",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				filepath.Join("a5", "48", "92", "a54892"): {
					mockDirEntry{name: "a54892.txt", isDir: false},
					mockDirEntry{name: ".hidden.txt", isDir: false},
					mockDirEntry{name: ".hidden", isDir: true},
				},
				filepath.Join("a5", "48", "92", "a54892", ".hidden"): {mockDirEntry{name: "innerHidden.txt", isDir: false}},
			},
		},
		{
			pairpath:    filepath.Join("b5", "48", "8", "b5488"),
			id:          "b5488",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				filepath.Join("b5", "48", "8", "b5488"): {
					mockDirEntry{name: "outerb5488.txt", isDir: false},
					mockDirEntry{name: "folder", isDir: true},
				},
				filepath.Join("b5", "48", "8", "b5488", "folder"): {
					mockDirEntry{name: ".hiddenFile.txt", isDir: false},
					mockDirEntry{name: "innerb5488.txt", isDir: false},
					mockDirEntry{name: ".hidden", isDir: true},
				},
				filepath.Join("b5", "48", "8", "b5488", "folder", ".hidden"): {
					mockDirEntry{name: "inner.txt", isDir: false},
				},
			},
		},
		{
			pairpath:    "doesNotExist",
			id:          "doesNotExist",
			expectError: os.ErrNotExist,
			expectMap:   nil,
		},
	}

	fs := afero.NewOsFs()

	for _, test := range tests {
		t.Run(test.pairpath, func(t *testing.T) {
			// Create a temporary directory for this test
			tempDir := testutils.CreateTempDir(t, fs)

			testutils.CopyTestDirectory(t, testutils.TestPairtree, tempDir)

			// Create the new testpath that has the full directory name
			prefixPairtree := filepath.Join(tempDir, rootDir)
			updatedMap := updateMapKeys(test.expectMap, prefixPairtree)
			fullPath := filepath.Join(prefixPairtree, test.pairpath)
			resultMap, err := RecursiveFiles(fullPath, test.id)
			// Compare actual results with the expected results
			assert.ErrorIs(t, err, test.expectError)
			assert.True(t, CompareMaps(updatedMap, resultMap), "Expected map: %v, Got: %v", updatedMap, resultMap)
		})
	}
}

// TestGetPrefix creates a temporary directory with Afero and alters the prefix file depending on test needs
func TestNonRecursiveFiles(t *testing.T) {
	tests := []struct {
		pairpath    string
		id          string
		expectError error
		expectMap   map[string][]fs.DirEntry
	}{
		{
			pairpath:    filepath.Join("a5", "38", "8", "a5388"),
			id:          "a5388",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				filepath.Join("a5", "38", "8", "a5388"): {
					mockDirEntry{name: "a5388.txt", isDir: false},
				},
			},
		},
		{
			pairpath:    filepath.Join("a5", "48", "92", "a54892"),
			id:          "a54892",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				filepath.Join("a5", "48", "92", "a54892"): {
					mockDirEntry{name: "a54892.txt", isDir: false},
					mockDirEntry{name: ".hidden.txt", isDir: false},
					mockDirEntry{name: ".hidden", isDir: true},
				},
			},
		},
		{
			pairpath:    filepath.Join("b5", "48", "8", "b5488"),
			id:          "b5488",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				filepath.Join("b5", "48", "8", "b5488"): {
					mockDirEntry{name: "outerb5488.txt", isDir: false},
					mockDirEntry{name: "folder", isDir: true},
				},
			},
		},
		{
			pairpath:    "doesNotExist",
			id:          "doesNotExist",
			expectError: os.ErrNotExist,
			expectMap:   nil,
		},
	}

	fs := afero.NewOsFs()

	for _, test := range tests {
		t.Run(test.pairpath, func(t *testing.T) {
			// Create a temporary directory for this test
			tempDir := testutils.CreateTempDir(t, fs)

			testutils.CopyTestDirectory(t, testutils.TestPairtree, tempDir)
			// Create the new testpath that has the full directory name
			prefixPairtree := filepath.Join(tempDir, rootDir)
			updatedMap := updateMapKeys(test.expectMap, prefixPairtree)
			fullPath := filepath.Join(prefixPairtree, test.pairpath)
			resultMap, err := NonRecursiveFiles(fullPath)
			// Compare actual results with the expected results
			assert.ErrorIs(t, err, test.expectError)
			assert.True(t, CompareMaps(updatedMap, resultMap), "Expected map: %v, Got: %v", updatedMap, resultMap)
		})
	}
}

func TestCheckPTVer(t *testing.T) {
	tests := []struct {
		name      string
		expectErr error
	}{
		{
			name:      "noVerFile",
			expectErr: os.ErrNotExist,
		},
		{
			name:      "verFileExist",
			expectErr: nil,
		},
		{
			name:      "verFileEmpty",
			expectErr: error_msgs.Err2,
		},
	}
	fs := afero.NewOsFs()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a temporary directory for this test
			tempDir := testutils.CreateTempDir(t, fs)

			// testPath := filepath.Join(tempDir, test.name)
			err := copy.Copy(testutils.TestPairtree, tempDir)
			if err != nil {
				t.Fatalf("Error copying directory: %v", err)
			}
			verFile := filepath.Join(tempDir, verDir)

			if test.name == "noVerFile" {
				err = fs.Remove(verFile)
				if err != nil {
					t.Errorf("Error deleting file %s: %v", verFile, err)
				}
			} else if test.name == "verFileEmpty" {
				err = afero.WriteFile(fs, verFile, []byte{}, 0644)
				if err != nil {
					t.Errorf("Error clearing file %s: %v", verFile, err)
				}
			}

			err = CheckPTVer(tempDir)
			assert.ErrorIs(t, err, test.expectErr)

		})
	}

}

func TestCreateDirNotExist(t *testing.T) {
	// Define an in-memory filesystem using afero
	fs := afero.NewOsFs()

	// Define test cases
	tests := []struct {
		name     string
		path     string
		setup    func(afero.Fs, string) error
		expected error
	}{
		{
			name: "directory does not exist",
			path: "testdir_not_exist",
			setup: func(fs afero.Fs, path string) error {
				// Ensure the directory does not exist before the test
				return fs.RemoveAll(path)
			},
			expected: nil,
		},
		{
			name: "directory already exists",
			path: "testdir_exist",
			setup: func(fs afero.Fs, path string) error {
				// Create the directory before the test
				return fs.MkdirAll(path, 0755)
			},
			expected: nil,
		},
		{
			name: "path is not allowed",
			path: "",
			setup: func(fs afero.Fs, path string) error {
				return nil
			},
			expected: error_msgs.Err15,
		},
	}

	// Run each test case
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Perform the setup for each test case
			if err := test.setup(fs, test.path); err != nil {
				t.Fatalf("Failed setup: %v", err)
			}

			// Call the function under test
			err := CreateDirNotExist(test.path)

			// Check the result
			if test.expected != nil {
				assert.Error(t, err)
				assert.IsType(t, test.expected, err)
			} else {
				assert.NoError(t, err)

				// Verify the directory was created if it did not exist
				exists, statErr := afero.DirExists(fs, test.path)
				if statErr != nil || !exists {
					t.Errorf("expected directory to exist: %s", test.path)
				}
			}
		})
	}
}

// TestCreatePairtree tests the CreatePairtree function with no prefix and a prefix provided
func TestCreatePairtree(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		path     string
		prefix   string
		expected error
	}{
		{
			name:     "no prefix",
			path:     "testdir",
			prefix:   "",
			expected: nil,
		},
		{
			name:     "has prefix",
			path:     "testDir",
			prefix:   prefix,
			expected: nil,
		},
		{
			name:     "has subdirectories",
			path:     filepath.Join("directory", "subdirectory"),
			prefix:   prefix,
			expected: nil,
		},
		{
			name:     "path is not allowed",
			path:     "  ",
			prefix:   "",
			expected: error_msgs.Err15,
		},
	}

	fs := afero.NewOsFs()

	// Run each test case
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err error
			var tempDir string

			if strings.TrimSpace(test.path) == "" {
				tempDir = test.path
			} else {
				tempDir = testutils.CreateTempDir(t, fs)
				tempDir = filepath.Join(tempDir, test.path)
			}

			err = CreatePairtree(tempDir, prefix)
			require.ErrorIs(t, err, test.expected)

			if test.expected == nil {
				ptPreFilePath := filepath.Join(tempDir, prefixDir)
				ptVerFilePath := filepath.Join(tempDir, verDir)
				ptRootDirPath := filepath.Join(tempDir, rootDir)

				// check prefix
				ptPre, err := testutils.OpenFileAndCheck(fs, ptPreFilePath)
				assert.ErrorIs(t, err, nil, "There was an error opening the prefix file")
				ptPreStirng := string(ptPre)
				assert.Equal(t, prefix, ptPreStirng, "The prefix in the file did not match the prefix given to CreatePairtree()")

				// check version
				ptVerContent, err := testutils.OpenFileAndCheck(fs, ptVerFilePath)
				assert.ErrorIs(t, err, nil, "There was an error opening the prefix file")
				ptVerString := string(ptVerContent)
				assert.Equal(t, ptVerSpec, ptVerString, "The version in the file did not match the expected version")
				//check if the directory was created

				// Use os.Stat to get the file info for the path
				info, err := os.Stat(ptRootDirPath)
				assert.ErrorIs(t, err, nil, "There was an error with creating the pt_root dir")
				assert.True(t, info.IsDir(), "The pt_root is not appearing as a directory")
			}
		})
	}
}

// TestBuildDirectoryTree tests the BuildDirectoryTree function
func TestBuildDirectoryTree(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		entriesMap       map[string][]fs.DirEntry
		isFirstIteration bool
		expected         Directory
	}{
		{
			name: "SimpleDirectoryStructure",
			path: filepath.Join("root"),
			entriesMap: map[string][]fs.DirEntry{
				filepath.Join("root"): {
					mockDirEntry{name: "file1.txt", isDir: false},
					mockDirEntry{name: "dir1", isDir: true},
				},
				filepath.Join("root", "dir1"): {
					mockDirEntry{name: "file2.txt", isDir: false},
				},
			},
			isFirstIteration: true,
			expected: Directory{
				Name: filepath.Join("root"),
				Directories: []Directory{
					{
						Name: "dir1",
						Files: []File{
							{Name: "file2.txt"},
						},
					},
				},
				Files: []File{
					{Name: "file1.txt"},
				},
			},
		},
		{
			name: "EmptyDirectory",
			path: filepath.Join("root"),
			entriesMap: map[string][]fs.DirEntry{
				filepath.Join("root"): {},
			},
			isFirstIteration: true,
			expected: Directory{
				Name: filepath.Join("root"),
			},
		},
		{
			name: "NestedDirectories",
			path: filepath.Join("root"),
			entriesMap: map[string][]fs.DirEntry{
				filepath.Join("root"): {
					mockDirEntry{name: "dir1", isDir: true},
				},
				filepath.Join("root", "dir1"): {
					mockDirEntry{name: "dir2", isDir: true},
				},
				filepath.Join("root", "dir1", "dir2"): {
					mockDirEntry{name: "file1.txt", isDir: false},
				},
			},
			isFirstIteration: true,
			expected: Directory{
				Name: filepath.Join("root"),
				Directories: []Directory{
					{
						Name: "dir1",
						Directories: []Directory{
							{
								Name: "dir2",
								Files: []File{
									{Name: "file1.txt"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "NestedDirWFiles",
			path: filepath.Join("root"),
			entriesMap: map[string][]fs.DirEntry{
				filepath.Join("root"): {
					mockDirEntry{name: "dir1", isDir: true},
					mockDirEntry{name: "file1.txt", isDir: false},
					mockDirEntry{name: "file2.txt", isDir: false},
				},
				filepath.Join("root", "dir1"): {
					mockDirEntry{name: "dir2", isDir: true},
				},
				filepath.Join("root", "dir1", "dir2"): {
					mockDirEntry{name: "file3.txt", isDir: false},
				},
			},
			isFirstIteration: true,
			expected: Directory{
				Name: filepath.Join("root"),
				Directories: []Directory{
					{
						Name: "dir1",
						Directories: []Directory{
							{
								Name: "dir2",
								Files: []File{
									{Name: "file3.txt"},
								},
							},
						},
					},
				},
				Files: []File{
					{Name: "file1.txt"},
					{Name: "file2.txt"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := BuildDirectoryTree(test.path, test.entriesMap, test.isFirstIteration)
			assert.True(t, compareDirectories(result, test.expected), "Expected map %+v, got %+v", test.expected, result)

		})
	}
}

// TestToJSONStructure tests the function that turns a directory map into a json structure
func TestToJSONStructure(t *testing.T) {
	tests := []struct {
		name       string
		dirTree    Directory
		expectJSON string
		expectErr  error
	}{
		{
			name: "empty directory",
			dirTree: Directory{
				Name:        "root",
				Directories: []Directory{},
				Files:       []File{},
			},
			expectJSON: `{
			"name": "root",
			"directories": [],
			"files": []
			}`,
			expectErr: nil,
		},
		{
			name: "directory with files",
			dirTree: Directory{
				Name: "root",
				Directories: []Directory{
					{
						Name:        "subdir",
						Directories: []Directory{},
						Files:       []File{},
					},
				},
				Files: []File{
					{Name: "file1.txt"},
					{Name: "file2.txt"},
				},
			},
			expectJSON: `{
			"name": "root",
			"directories": [
				{
				"name": "subdir",
				"directories": [],
				"files": []
				}
			],
			"files": [
				{
				"name": "file1.txt"
				},
				{
				"name": "file2.txt"
				}
			]
			}`,
			expectErr: nil,
		},
		{
			name: "nested directories",
			dirTree: Directory{
				Name: "root",
				Directories: []Directory{
					{
						Name: "subdir1",
						Directories: []Directory{
							{
								Name:        "subsubdir1",
								Directories: []Directory{},
								Files:       []File{{Name: "file3.txt"}},
							},
						},
						Files: []File{},
					},
				},
				Files: []File{},
			},
			expectJSON: `{
			"name": "root",
			"directories": [
				{
				"name": "subdir1",
				"directories": [
					{
					"name": "subsubdir1",
					"directories": [],
					"files": [
						{
						"name": "file3.txt"
						}
					]
					}
				],
				"files": []
				}
			],
			"files": []
			}`,
			expectErr: nil,
		},
		// Add more test cases as needed
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotJSON, err := ToJSONStructure(test.dirTree)
			assert.True(t, errors.Is(err, test.expectErr))

			assert.JSONEq(t, test.expectJSON, string(gotJSON))

		})
	}
}

// TestDeletePairtreeItem tests if directories and files are deleted when passed in
func TestDeletePairtreeItem(t *testing.T) {
	tests := []struct {
		name        string
		pairpath    string
		expectError error
	}{
		{
			name:        "file",
			pairpath:    filepath.Join("a5", "38", "8", "a5388", "a5388.txt"),
			expectError: nil,
		},
		{
			name:        "directory",
			pairpath:    filepath.Join("b5", "48", "8", "b5488", "folder"),
			expectError: nil,
		},
		{
			name:        "object",
			pairpath:    filepath.Join("a5", "48", "8", "a5488"),
			expectError: nil,
		},
		{
			name:        "doesNotExist",
			pairpath:    "doesNotExist",
			expectError: os.ErrNotExist,
		},
	}

	fs := afero.NewOsFs()

	for _, test := range tests {
		t.Run(test.pairpath, func(t *testing.T) {
			// Create a temporary directory for this test
			tempDir := testutils.CreateTempDir(t, fs)

			testutils.CopyTestDirectory(t, testutils.TestPairtree, tempDir)
			// Create the new testpath that has the full directory name
			prefixPairtree := filepath.Join(tempDir, rootDir)
			fullPath := filepath.Join(prefixPairtree, test.pairpath)
			err := DeletePairtreeItem(fullPath)
			// Compare actual results with the expected results
			assert.ErrorIs(t, err, test.expectError)
		})
	}
}

// TestCopyFile tests copying files into directories
func TestCopyFile(t *testing.T) {

	testFiles := []struct {
		testName       string
		fileName       string
		changeFileName bool
		overwrite      bool
		createDest     bool
		expectError    error
	}{
		{
			testName:       "No overwrite and change file name",
			fileName:       "newfilename",
			changeFileName: true,
			overwrite:      true,
			expectError:    nil,
		},
		{
			testName:       "No overwrite and same file name",
			fileName:       "",
			changeFileName: false,
			overwrite:      true,
			expectError:    nil,
		},
		{
			testName:       "Overwrite existing file",
			fileName:       ".1",
			changeFileName: false,
			overwrite:      false,
			expectError:    nil,
		},
	}

	fs := afero.NewOsFs()

	for _, test := range testFiles {
		t.Run(test.testName, func(t *testing.T) {

			destFilePath := ""
			content := []byte("File contents")
			tempFilePath := testutils.CreateTempFile(t, fs, content)
			tempFile := filepath.Base(tempFilePath)
			dirDest := testutils.CreateTempDir(t, fs)

			if test.changeFileName {
				destFilePath = filepath.Join(dirDest, test.fileName)
				dirDest = filepath.Join(dirDest, test.fileName)
			} else {
				// Build the expected destination file path
				destFilePath = filepath.Join(dirDest, tempFile)
			}

			_, err := CopyFileOrFolder(tempFilePath, dirDest, test.overwrite)
			assert.ErrorIs(t, err, test.expectError)

			// if the .x naming convetion should be used, recopy the file
			if !test.overwrite {
				_, err = CopyFileOrFolder(tempFilePath, dirDest, test.overwrite)
				assert.ErrorIs(t, err, test.expectError)
				destFilePath = destFilePath + test.fileName
			}

			// Check if the destination file exists
			exists, err := afero.Exists(fs, destFilePath)
			assert.ErrorIs(t, err, nil, "Failed to check if dirSrc was copied: %v", err)
			assert.True(t, exists, "File was not copied to destination")

			// Verify the copied file's contents match the original
			copiedFileContent, err := afero.ReadFile(fs, destFilePath)
			if err != nil {
				t.Fatalf("Failed to read copied file: %v", err)
			}
			assert.Equal(t, content, copiedFileContent, "Copied file content does not match the original")

			if test.changeFileName || !test.overwrite {
				assert.NotEqual(t, filepath.Base(tempFilePath), filepath.Base(destFilePath), "File names match and should not")
			} else {
				// Check that the file name matches
				assert.Equal(t, filepath.Base(tempFilePath), filepath.Base(destFilePath), "File names do not match")
			}
		})
	}
}

// TestCopyFolder tests copying a directory into another directory
func TestCopyFolder(t *testing.T) {
	testFolders := []struct {
		testName         string
		folderName       string
		changeFolderName bool
		overwrite        bool
		expectError      error
		expFoldName      string
	}{
		{
			testName:         "Basic copy of folder",
			folderName:       "folderExists",
			changeFolderName: false,
			overwrite:        true,
			expectError:      nil,
			expFoldName:      filepath.Join("folderExists", "folder"),
		},
		{
			testName:         "Slash added to folder name",
			folderName:       "folderExists" + string(os.PathSeparator),
			changeFolderName: false,
			overwrite:        true,
			expectError:      nil,
			expFoldName:      filepath.Join("folderExists", "folder"),
		},
		{
			testName:         "New folder name",
			folderName:       "newFolder",
			changeFolderName: true,
			overwrite:        true,
			expectError:      nil,
			expFoldName:      filepath.Join("newFolder"),
		},
		{
			testName:         "Do not overwrite folder and use .x",
			folderName:       "noOverwrite",
			changeFolderName: false,
			overwrite:        false,
			expectError:      nil,
			expFoldName:      filepath.Join("noOverwrite", "folder.1"),
		},
	}

	fs := afero.NewOsFs()

	for _, test := range testFolders {
		t.Run(test.testName, func(t *testing.T) {
			srcFolder := "folder"

			dirSrc := testutils.CreateTempDir(t, fs)
			dirDest := testutils.CreateTempDir(t, fs)

			dirSrc = testutils.CreateDirInDir(t, fs, dirSrc, srcFolder)
			_ = testutils.CreateFileInDir(t, dirSrc, "file.txt")

			if test.changeFolderName {
				dirDest = filepath.Join(dirDest, test.folderName)
			} else {
				dirDest = testutils.CreateDirInDir(t, fs, dirDest, test.folderName)
			}

			if strings.HasSuffix(test.folderName, string(os.PathSeparator)) {
				dirDest += string(os.PathSeparator)
			}

			finalDest, err := CopyFileOrFolder(dirSrc, dirDest, test.overwrite)
			assert.ErrorIs(t, err, test.expectError, "Expected CopyFilrOrFolder to return %v", err)

			if !test.overwrite {
				finalDest, err = CopyFileOrFolder(dirSrc, dirDest, test.overwrite)
				assert.ErrorIs(t, err, test.expectError)
			}
			exists, err := afero.DirExists(fs, finalDest)
			assert.ErrorIs(t, err, nil, "Failed to check if dirSrc was copied: %v", err)

			// Optionally, check if the contents of dirSrc were copied
			assert.True(t, exists, "Directory %s was not copied to destination: %s", dirSrc, dirDest)

			srcFiles, err := afero.ReadDir(fs, dirSrc)
			assert.ErrorIs(t, err, nil, "Failed to read dirSrc contents: %v", err)

			assert.True(t, strings.HasSuffix(finalDest, test.expFoldName), "Folder path should have %s", test.expFoldName)

			destFiles, err := afero.ReadDir(fs, finalDest)
			assert.ErrorIs(t, err, nil, "Failed to read copied dir contents: %v", err)

			// Ensure that the number of files and directories match
			assert.Equal(t, len(srcFiles), len(destFiles), "Number of files in source and destination do not match")

			// Further checks can compare individual file names and contents
			for i, srcFile := range srcFiles {
				assert.Equal(t, srcFile.Name(), destFiles[i].Name(), "File names do not match")
			}
		})
	}

}

// TestGetUniqueDestinationTabular runs tabular tests for the GetUniqueDestination function
func TestGetUniqueDestination(t *testing.T) {
	// Define the test cases
	tests := []struct {
		name           string
		existingFiles  []string // Files that already exist in the destination
		expectedSuffix string   // Expected suffix for the unique file
	}{
		{
			name:           "No Existing File",
			existingFiles:  []string{}, // No existing files
			expectedSuffix: "",         // Should return the original name
		},
		{
			name:           "Single Existing File",
			existingFiles:  []string{"file.txt"}, // One file exists
			expectedSuffix: ".1",                 // Should return file.1.txt
		},
		{
			name:           "Multiple Existing Files",
			existingFiles:  []string{"file.txt", "file.1.txt", "file.2.txt"}, // Multiple files exist
			expectedSuffix: ".3",                                             // Should return file.3.txt
		},
		{
			name:           "Non-Conflicting File",
			existingFiles:  []string{"otherfile.txt"}, // Different file exists, no conflict
			expectedSuffix: "",                        // Should return the original name
		},
	}

	fs := afero.NewOsFs()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a temporary directory
			tempDir := testutils.CreateTempDir(t, fs)

			// Define the destination file path
			destPath := filepath.Join(tempDir, "file.txt")

			// Create any existing files needed for the test
			for _, file := range test.existingFiles {
				existingFilePath := filepath.Join(tempDir, file)
				err := afero.WriteFile(fs, existingFilePath, []byte("existing content"), 0644)
				assert.NoError(t, err, "Failed to create existing file: %s", file)
			}

			// Call the function under test
			uniquePath := GetUniqueDestination(destPath)

			// Calculate the expected unique path
			expectedPath := filepath.Join(tempDir, "file"+test.expectedSuffix+".txt")

			// Verify the result
			assert.Equal(t, expectedPath, uniquePath, "Unique path mismatch for test case: %s", test.name)
		})
	}
}

// TestTarGz tests the TarGz function with different test cases using tabular testing and afero.
func TestTarGz(t *testing.T) {
	// Test cases for the TarGz function
	tests := []struct {
		name       string
		prefix     string
		encodedPre string
		overwrite  bool
		expectErr  error
	}{
		{
			name:       "No prefix new TarGz Archive",
			prefix:     "",
			encodedPre: "",
			overwrite:  true,
			expectErr:  nil,
		},
		{
			name:       "Prefix new TarGz Archive",
			prefix:     "ark:/",
			encodedPre: "ark+=",
			overwrite:  true,
			expectErr:  nil,
		},
		{
			name:       "No overwrite or prefix",
			prefix:     "",
			encodedPre: "",
			overwrite:  false,
			expectErr:  nil,
		},
		{
			name:       "No overwrite with prefix",
			prefix:     "ark:/",
			encodedPre: "ark+=",
			overwrite:  false,
			expectErr:  nil,
		},
	}
	// Create an afero in-memory filesystem
	fs := afero.NewOsFs()

	// Loop through each test case
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dirSrc := testutils.CreateTempDir(t, fs)
			dirDest := testutils.CreateTempDir(t, fs)

			_ = testutils.CreateFileInDir(t, dirSrc, "file.txt")

			// Call the TarGz function
			err := TarGz(dirSrc, dirDest, test.prefix, test.overwrite)
			assert.ErrorIs(t, err, test.expectErr, "There was an Error with TarGZ")

			tarDest := filepath.Join(dirDest, test.encodedPre+filepath.Base(dirSrc)+".tgz")

			// Check if overwrite behavior was respected
			if !test.overwrite {
				err = TarGz(dirSrc, dirDest, test.prefix, test.overwrite)
				assert.ErrorIs(t, err, test.expectErr, "There was an Error with TarGZ")

				tarDest = filepath.Join(dirDest, test.encodedPre+filepath.Base(dirSrc)+".1"+".tgz")
			}
			// Check if the tar.gz file was created in the destination directory
			exists, err := afero.Exists(fs, tarDest)
			assert.NoError(t, err, "error checking for tar.gz file existence")
			assert.True(t, exists, ".tgz file does not exist")
		})
	}
}

func TestUnTarGz(t *testing.T) {
	tests := []struct {
		name      string
		addFolder bool
		srcID     string
		tgzID     string
		expectErr error
	}{
		{
			name:      "Untar file properly",
			addFolder: false,
			srcID:     "folderID",
			tgzID:     "folderID",
			expectErr: nil,
		},
		{
			name:      "Folder in .tgz does not match src folder",
			addFolder: false,
			srcID:     "folderID",
			tgzID:     "folderIDNotMatch",
			expectErr: error_msgs.Err13,
		},
		{
			name:      "More than one folder exists in .tgz",
			addFolder: true,
			srcID:     "folderID",
			tgzID:     "folderID",
			expectErr: error_msgs.Err12,
		},
	}

	// Create an afero in-memory filesystem
	fs := afero.NewOsFs()

	// Loop through each test case
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			dirDest := testutils.CreateTempDir(t, fs)
			dirDest = testutils.CreateDirInDir(t, fs, dirDest, test.srcID)

			//Create the .tgz in a temporary directory
			tempDir := testutils.CreateTempDir(t, fs)
			dirTGZ := testutils.CreateDirInDir(t, fs, tempDir, test.tgzID)

			dirSrcTGZ := filepath.Join(tempDir, test.tgzID+".tgz")

			fileNames := []string{"file.txt", "file1.txt", "file2.txt"}
			for _, fileName := range fileNames {
				_ = testutils.CreateFileInDir(t, dirTGZ, fileName)
			}
			sourceFolders := []string{dirTGZ}

			if test.addFolder {
				pathToFolder := testutils.CreateDirInDir(t, fs, tempDir, "extraFolder")
				sourceFolders = append(sourceFolders, pathToFolder)
			}

			tgz := archiver.NewTarGz()

			// Archive the source directory
			if err := tgz.Archive(sourceFolders, dirSrcTGZ); err != nil {
				t.Fatalf("There was an error archiving the folder %v", err)
			}
			err := UnTarGz(dirSrcTGZ, dirDest)

			assert.ErrorIs(t, err, test.expectErr)
		})
	}
}
