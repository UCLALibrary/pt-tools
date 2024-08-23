package ptls

/*ptls: an ls-like tool that can display the contents of the Pairtree object; options 
include: -a (but have this work like ls' -A which does include the . and .. directories in the
output), -d (which only lists directories of the object directory), -j (which returns output in a
JSON structure instead of basic string output), and -R (for a recursive listing of the object directory,
with the default being a non-recursive listing). The basic command is ptls [ID]
(when an ENV PAIRTREE_ROOT is set) or ptls [PT_ROOT] [ID]) with the output listing the contents of
the Pairtree object directory (doing all the navigation through the Pairtree structure behind the scenes).
It also supports -h for details about what it can do.*/

// Just one ID
import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	"github.com/UCLALibrary/pt-tools/pkg/pairtree"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// FileInfo holds the name and type of a directory entry.
type FileInfo struct {
	Path     string
	IsDir    bool
	IsHidden bool
}

var (
	showAll      bool
	showDirsOnly bool
	outputJSON   bool
	recursive    bool
	ptRoot       string
	Logger       *zap.Logger = logger()
	logFile      string      = "logs.log"
	id           string      = ""
)

// ApplyExitOnHelp exits out of program if --help is flag
func ApplyExitOnHelp(c *cobra.Command, exitCode int) {
	helpFunc := c.HelpFunc()
	c.SetHelpFunc(func(c *cobra.Command, s []string) {
		helpFunc(c, s)
		os.Exit(exitCode)
	})
}

// logger creates logger with output of info and debug to file and error to stdout
func logger() *zap.Logger {
	pe := zap.NewDevelopmentEncoderConfig()

	fileEncoder := zapcore.NewJSONEncoder(pe)

	pe.EncodeTime = zapcore.ISO8601TimeEncoder // The encoder can be customized for each output

	// Console encoder (for stdout)
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	// Create file core
	file, err := os.Create(logFile)
	if err != nil {
		panic(err)
	}

	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(file), zap.DebugLevel)

	// Console core for errors
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.ErrorLevel)

	// Combine the cores
	core := zapcore.NewTee(fileCore, consoleCore)
	// Create a logger with two cores
	logger := zap.New(core, zap.AddCaller())

	return logger
}

func initFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&showAll, "a", "a", false, "do not ignore entries starting with .")
	cmd.Flags().BoolVarP(&showDirsOnly, "d", "d", false, "list directories only")
	cmd.Flags().BoolVarP(&outputJSON, "j", "j", false, "output in JSON format")
	cmd.Flags().BoolVarP(&recursive, "r", "r", false, "list directories recursively")
	cmd.Flags().StringVarP(&ptRoot, "Root", "R", "", "Set root directory")

}

func Run(args []string, writer io.Writer) error {
	var ptMap map[string][]fs.DirEntry
	var err error
	var pairPath string

	var rootCmd = &cobra.Command{
		Use:   "ptls [PT_ROOT] [FLAGS] [ID]",
		Short: "ptls is a tool to list Pairtree object directories.",
		Long:  "A tool to list contents of Pairtree object directories with various options.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				fmt.Fprintln(writer, "Please provide an ID for the pairtree")
				Logger.Error("Error getting ID",
					zap.Error(error_msgs.Err6))

				return error_msgs.Err6
			}
			// Extract the ID from the final argument
			id = args[len(args)-1]

			// If the root has not been set yet check the ENV vars
			if ptRoot == "" {

				if envVar := os.Getenv("PAIRTREE_ROOT"); envVar != "" {
					ptRoot = envVar
				} else {
					fmt.Fprintln(writer, error_msgs.Err7)
					return error_msgs.Err7
				}
			}

			Logger.Info("Pairtree root is",
				zap.String("PAIRTREE_ROOT", ptRoot),
			)
			return nil
		},
	}

	initFlags(rootCmd)
	rootCmd.SetOut(writer)
	rootCmd.SetErr(writer)
	rootCmd.SetArgs(args)

	ApplyExitOnHelp(rootCmd, 0)

	if err = rootCmd.Execute(); err != nil {
		Logger.Error("Error setting command line",
			zap.Error(err))
		return err
	}

	// check if the pairtree version file exists and is populated
	if err := pairtree.CheckPTVer(ptRoot); err != nil {
		Logger.Error("Error with pairtree veresion file", zap.Error(err))
		return err
	}

	// Get the prefix from pairtree_prefix file
	prefix, err := pairtree.GetPrefix(ptRoot)

	if err != nil {
		Logger.Error("Error retrieving prefix from pairtree_prefix file", zap.Error(err))
		return err
	}

	// create the pairpath
	pairPath, err = pairtree.CreatePP(id, ptRoot, prefix)

	if err != nil {
		Logger.Error("Error creating pairpath", zap.Error(err))
		return err
	}

	if recursive {
		ptMap, err = pairtree.RecursiveFiles(pairPath, id)
		if err != nil {
			Logger.Error("Error retrieving list of files recursively", zap.Error(err))
			return err
		}
	} else {
		ptMap, err = pairtree.NonRecursiveFiles(pairPath)
		if err != nil {
			Logger.Error("Error retrieving list of files recursively", zap.Error(err))
			return err
		}
	}

	if showDirsOnly {
		// Filter ptMap to only include directories
		for key, entries := range ptMap {
			var filteredEntries []fs.DirEntry
			for _, entry := range entries {
				if pairtree.IsDirectory(entry) {
					filteredEntries = append(filteredEntries, entry)
				}
			}
			if len(filteredEntries) > 0 {
				ptMap[key] = filteredEntries
			} else {
				delete(ptMap, key)
			}
		}
	}

	// If hidden files and directories should be removed from the map
	if !showAll {
		for key, entries := range ptMap {
			// Check if the key (directory name) is hidden
			if pairtree.IsHidden(filepath.Base(key)) {
				// If the key is hidden, remove it from the map
				delete(ptMap, key)
				continue
			}

			// Filter out hidden entries within the directory
			var filteredEntries []fs.DirEntry
			for _, entry := range entries {
				if !pairtree.IsHidden(entry.Name()) {
					filteredEntries = append(filteredEntries, entry)
				}
			}

			// Update the map with filtered entries or remove the key if no entries remain
			if len(filteredEntries) > 0 {
				ptMap[key] = filteredEntries
			} else {
				delete(ptMap, key)
			}
		}
	}

	if outputJSON {
		dirTree := pairtree.BuildDirectoryTree(pairPath, ptMap, true)

		recursiveJSON, err := pairtree.ToJSONStructure(pairPath, dirTree)
		if err != nil {
			Logger.Error("Error converting to Json", zap.Error(err))
			return err
		}
		fmt.Fprintf(writer, "JSON structure:\n%s\n", string(recursiveJSON))
	} else {

		// Display the directory structure
		for dir, entries := range ptMap {
			fmt.Fprintln(writer, dir+":")
			for _, entry := range entries {
				if pairtree.IsDirectory(entry) {
					fmt.Fprintf(writer, "  %s/\n", entry.Name())
				} else {
					fmt.Fprintf(writer, "  %s\n", entry.Name())
				}
			}
		}

	}

	return nil
}
