package pairtree

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	"github.com/UCLALibrary/pt-tools/pkg/testutils"
	"github.com/otiai10/copy"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	testPairtree = "../../test-dir/test-pairtree"
	prefix       = "ark:/"
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

// CreateFileInDir creates a file with the given name in the specified directory.
func CreateFileInDir(dir string, filename string) error {
	// Join the directory and filename to get the full path of the new file.
	filePath := filepath.Join(dir, filename)

	// Create the file.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	return nil
}

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
			expectPre:   "ark:/",
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

			// Copies entire directory in testPairTree into the temporary directory
			testutils.CopyTestDirectory(t, testPairtree, tempDir)

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
		expectPP  string
	}{
		{
			name:      "standard",
			id:        "ark:/345621",
			ptRoot:    "root",
			prefix:    "ark:/",
			expectErr: nil,
			expectPP:  "root/pairtree_root/34/56/21/345621",
		},
		{
			name:      "specialChars",
			id:        "ark:/34:621",
			ptRoot:    "root",
			prefix:    "ark:/",
			expectErr: nil,
			expectPP:  "root/pairtree_root/34/+6/21/34+621",
		},
		{
			name:      "noPrefix",
			id:        "34621",
			ptRoot:    "root",
			prefix:    "",
			expectErr: nil,
			expectPP:  "root/pairtree_root/34/62/1/34621",
		},
		{
			name:      "noPtRoot",
			id:        "34621",
			ptRoot:    "",
			prefix:    "ark:/",
			expectErr: error_msgs.Err3,
			expectPP:  "",
		},
		{
			name:      "noId",
			id:        "",
			ptRoot:    "root",
			prefix:    "",
			expectErr: error_msgs.Err4,
			expectPP:  "",
		},
		{
			name:      "idNoPrefix",
			id:        "34621",
			ptRoot:    "root",
			prefix:    "ark:/",
			expectErr: error_msgs.Err5,
			expectPP:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pairpath, err := CreatePP(test.id, test.ptRoot, test.prefix)
			assert.Equal(t, test.expectPP, pairpath)
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
			pairpath:    "a5/38/8/a5388",
			id:          "a5388",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				"a5/38/8/a5388": {
					mockDirEntry{name: "a5388.txt", isDir: false},
				},
			},
		},
		{
			pairpath:    "a5/48/92/a54892",
			id:          "a54892",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				"a5/48/92/a54892": {
					mockDirEntry{name: "a54892.txt", isDir: false},
					mockDirEntry{name: ".hidden.txt", isDir: false},
					mockDirEntry{name: ".hidden", isDir: true},
				},
				"a5/48/92/a54892/.hidden": {mockDirEntry{name: "innerHidden.txt", isDir: false}},
			},
		},
		{
			pairpath:    "b5/48/8/b5488",
			id:          "b5488",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				"b5/48/8/b5488": {
					mockDirEntry{name: "outerb5488.txt", isDir: false},
					mockDirEntry{name: "folder", isDir: true},
				},
				"b5/48/8/b5488/folder": {
					mockDirEntry{name: ".hiddenFile.txt", isDir: false},
					mockDirEntry{name: "innerb5488.txt", isDir: false},
					mockDirEntry{name: ".hidden", isDir: true},
				},
				"b5/48/8/b5488/folder/.hidden": {
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

			testutils.CopyTestDirectory(t, testPairtree, tempDir)

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
			pairpath:    "a5/38/8/a5388",
			id:          "a5388",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				"a5/38/8/a5388": {
					mockDirEntry{name: "a5388.txt", isDir: false},
				},
			},
		},
		{
			pairpath:    "a5/48/92/a54892",
			id:          "a54892",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				"a5/48/92/a54892": {
					mockDirEntry{name: "a54892.txt", isDir: false},
					mockDirEntry{name: ".hidden.txt", isDir: false},
					mockDirEntry{name: ".hidden", isDir: true},
				},
			},
		},
		{
			pairpath:    "b5/48/8/b5488",
			id:          "b5488",
			expectError: nil,
			expectMap: map[string][]fs.DirEntry{
				"b5/48/8/b5488": {
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

			testutils.CopyTestDirectory(t, testPairtree, tempDir)
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
			err := copy.Copy(testPairtree, tempDir)
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

// TestEncodeChar tests the EncodeChar function
func TestEncodeChar(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc", "abc"}, // No special characters
		{"ark:/13030/xt12t3", "ark+=13030=xt12t3"}, // Characters that should be replaced
		{"<>= ", "^3c^3e^3d^20"},                   // Characters that should be encoded
		{"Hello, World!", "Hello^2c^20World!"},     // Mixed case with some encoded characters
		{"ark://345", "ark+==345"},                 // Common file example
		{"http://n2t.info/urn:nbn:se:kb:repos-1", "http+==n2t,info=urn+nbn+se+kb+repos-1"},
		{"what-the-*@?#!^!?", "what-the-^2a@^3f#!^5e!^3f"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := EncodeChar(test.input)
			assert.Equal(t, test.expected, result)
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
			path: "/root",
			entriesMap: map[string][]fs.DirEntry{
				"/root": {
					mockDirEntry{name: "file1.txt", isDir: false},
					mockDirEntry{name: "dir1", isDir: true},
				},
				"/root/dir1": {
					mockDirEntry{name: "file2.txt", isDir: false},
				},
			},
			isFirstIteration: true,
			expected: Directory{
				Name: "/root",
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
			path: "/root",
			entriesMap: map[string][]fs.DirEntry{
				"/root": {},
			},
			isFirstIteration: true,
			expected: Directory{
				Name: "/root",
			},
		},
		{
			name: "NestedDirectories",
			path: "/root",
			entriesMap: map[string][]fs.DirEntry{
				"/root": {
					mockDirEntry{name: "dir1", isDir: true},
				},
				"/root/dir1": {
					mockDirEntry{name: "dir2", isDir: true},
				},
				"/root/dir1/dir2": {
					mockDirEntry{name: "file1.txt", isDir: false},
				},
			},
			isFirstIteration: true,
			expected: Directory{
				Name: "/root",
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
			path: "/root",
			entriesMap: map[string][]fs.DirEntry{
				"/root": {
					mockDirEntry{name: "dir1", isDir: true},
					mockDirEntry{name: "file1.txt", isDir: false},
					mockDirEntry{name: "file2.txt", isDir: false},
				},
				"/root/dir1": {
					mockDirEntry{name: "dir2", isDir: true},
				},
				"/root/dir1/dir2": {
					mockDirEntry{name: "file3.txt", isDir: false},
				},
			},
			isFirstIteration: true,
			expected: Directory{
				Name: "/root",
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
			assert.True(t, compareDirectories(result, test.expected), "Expected map %+v, got %+v", test.name, test.expected, result)

		})
	}
}

// TestToJSONStructure tests the function that turns a directory map into a json structure
func TestToJSONStructure(t *testing.T) {
	tests := []struct {
		name       string
		rootPath   string
		dirTree    Directory
		expectJSON string
		expectErr  error
	}{
		{
			name:     "empty directory",
			rootPath: "/",
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
			name:     "directory with files",
			rootPath: "/",
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
			name:     "nested directories",
			rootPath: "/",
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
			gotJSON, err := ToJSONStructure(test.rootPath, test.dirTree)
			assert.True(t, errors.Is(err, test.expectErr))

			assert.JSONEq(t, test.expectJSON, string(gotJSON))

		})
	}
}
