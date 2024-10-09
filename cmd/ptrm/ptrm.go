package ptrm

/*ptrm is a rm-like tool that can delete things from within a Pairtree object or
remove a Pairtree object altogether. There is also the ability to delete files and
directories in the object as long as the subpath to that file or directory is provided. */

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	"github.com/UCLALibrary/pt-tools/pkg/pairtree"
	"github.com/UCLALibrary/pt-tools/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// FileInfo holds the name and type of a directory entry.
type FileInfo struct {
	Path     string
	IsDir    bool
	IsHidden bool
}

var (
	ptRoot  string
	logFile string      = "logs.log"
	Logger  *zap.Logger = utils.Logger(logFile)
	id      string      = ""
	subpath string      = ""
)

func initFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&ptRoot, "pairtree", "p", "", "Set pairtree root directory")

}

func Run(args []string, writer io.Writer) error {
	var err error
	var pairPath string

	var rootCmd = &cobra.Command{
		Use:   "ptrm [PT_ROOT] [ID] [subpath/to/file.txt]",
		Short: "ptrm is a tool to remove Pairtree objects, files, and directores",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If the root has not been set yet check the ENV vars
			if ptRoot == "" {

				if envVar := os.Getenv("PAIRTREE_ROOT"); envVar != "" {
					ptRoot = envVar
				} else {
					fmt.Fprintln(writer, error_msgs.Err7)
					return error_msgs.Err7
				}
			}

			numArgs := len(args)
			if numArgs < 1 {
				fmt.Fprintln(writer, "Please provide an ID for the pairtree")
				Logger.Error("Error getting ID",
					zap.Error(error_msgs.Err6))

				return error_msgs.Err6
			}

			if numArgs == 1 {
				// Extract the ID from the final argument
				id = args[numArgs-1]
			} else if numArgs == 2 {
				// Extract the ID and the subpath from the arguments
				id = args[numArgs-2]
				subpath = args[numArgs-1]
			} else {
				fmt.Fprintln(writer, "Too many arguments were provided to ptrm")
				Logger.Error("Error parsing ptrm",
					zap.Error(error_msgs.Err8))

				return error_msgs.Err8
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

	utils.ApplyExitOnHelp(rootCmd, 0)

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

	if prefix == "" {
		prefix = pairtree.PtPrefix
	}
	// create the pairpath
	pairPath, err = pairtree.CreatePP(id, ptRoot, prefix)

	if err != nil {
		Logger.Error("Error creating pairpath", zap.Error(err))
		return err
	}

	fullPath := filepath.Join(pairPath, subpath)
	if err := pairtree.DeletePairtreeItem(fullPath); err != nil {
		Logger.Error("Error deleting pairpath", zap.Error(err))
		return err
	}

	fmt.Printf("Successfully deleted: %s\n", fullPath)

	return nil
}
