package ptls

/*ptls: an ls-like tool that can display the contents of the Pairtree object; options should
include: -a (but have this work like ls' -A which does include the . and .. directories in the
output), -d (which only lists directories of the object directory), -j (which returns output in a
JSON structure instead of basic string output), and -R (for a recursive listing of the object directory,
with the default being a non-recursive listing). The basic command should be: ptls [ID]
(when an ENV PAIRTREE_ROOT is set) or ptls [PT_ROOT] [ID]) with the output listing the contents of
the Pairtree object directory (doing all the navigation through the Pairtree structure behind the scenes).
This command (ptls) should also support trailing wildcards (not preceding wildcards), so things like:
ptls ark:/53355/cy88* should return the directories of multiple Pairtree objects (if multiples exist
	in the Pairtree). It should also support -h for details about what it can do.*/

// Just one ID
import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/UCLALibrary/pt-tools/pkg/pt"
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

var rootCmd = &cobra.Command{
	Use:   "ptls [PT_ROOT] [ID]",
	Short: "ptls is a tool to list Pairtree object directories.",
	Long:  "A tool to list contents of Pairtree object directories with various options.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Please provide an ID for the pairtree")
			os.Exit(1)
		}
		// Extract the ID from the final argument
		id = args[len(args)-1]

		// If the root has not been set yet check the ENV vars
		if ptRoot == "" {

			if envVar := os.Getenv("PAIRTREE_ROOT"); envVar != "" {
				ptRoot = envVar
			} else {
				fmt.Println(errors.New("--Root flag or PAIRTREE_ROOT environment variable must be set"))
				os.Exit(1)
			}
		}

		Logger.Info("Pairtree root is",
			zap.String("PAIRTREE_ROOT", ptRoot),
		)
	},
}

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

func init() {
	rootCmd.Flags().BoolVarP(&showAll, "a", "a", false, "do not ignore entries starting with .")
	rootCmd.Flags().BoolVarP(&showDirsOnly, "d", "d", false, "list directories only")
	rootCmd.Flags().BoolVarP(&outputJSON, "j", "j", false, "output in JSON format")
	rootCmd.Flags().BoolVarP(&recursive, "r", "r", false, "list directories recursively")
	rootCmd.Flags().StringVarP(&ptRoot, "Root", "R", "", "Set root directory")

}

func Run() error {
	var ptMap map[string][]fs.DirEntry
	var err error
	var pairPath string

	ApplyExitOnHelp(rootCmd, 0)

	if err = rootCmd.Execute(); err != nil {
		Logger.Error("Error setting command line",
			zap.Error(err))
		os.Exit(1)
	}

	pairPath, err = pt.CreatePP(id, ptRoot)

	if err != nil {
		Logger.Error("Error creating pairpath", zap.Error(err))
		os.Exit(1)
	}

	// Check if the pairtree version file exists and is populated
	if err := pt.CheckPTVer(ptRoot); err != nil {
		Logger.Error("Error with pairtree veresion file", zap.Error(err))
		os.Exit(1)
	}

	if recursive {
		ptMap, err = pt.RecursiveFiles(pairPath, id)
		if err != nil {
			Logger.Error("Error retrieving list of files recursively", zap.Error(err))
			os.Exit(1)
		}
	} else {
		ptMap, err = pt.NonRecursiveFiles(pairPath)
		if err != nil {
			Logger.Error("Error retrieving list of files recursively", zap.Error(err))
			os.Exit(1)
		}
	}

	if showDirsOnly {
		// Filter ptMap to only include directories
		for key, entries := range ptMap {
			var filteredEntries []fs.DirEntry
			for _, entry := range entries {
				if pt.IsDirectory(entry) {
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

	// If hidden files and dir should be removed from map
	if !showAll {
		for key, entries := range ptMap {
			var filteredEntries []fs.DirEntry
			for _, entry := range entries {
				if !pt.IsHidden(entry.Name()) {
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

	if outputJSON {
		recursiveJSON, err := pt.ToJSONStructure(pairPath, ptMap)
		if err != nil {
			Logger.Error("Error converting to Json", zap.Error(err))
			os.Exit(1)
		}
		fmt.Printf("Recursive JSON structure:\n%s\n", string(recursiveJSON))
	} else {

		// Display the directory structure
		for dir, entries := range ptMap {
			fmt.Println(dir + ":")
			for _, entry := range entries {
				if pt.IsDirectory(entry) {
					fmt.Printf("  %s/\n", entry.Name())
				} else {
					fmt.Printf("  %s\n", entry.Name())
				}
			}
			fmt.Println()
		}

	}

	return nil
}
