package ptnew

/* ptnew is a tool that creates the basic structure of a pairtree including the pairtree_version file, the pairtree_prefix file, and the pairtree_root folder */

import (
	"fmt"
	"io"
	"os"

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
	prefix  string
	logFile string      = "logs.log"
	Logger  *zap.Logger = utils.Logger(logFile)
)

func initFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&ptRoot, "pairtree", "p", "", "Set pairtree root directory")
	cmd.Flags().StringVarP(&prefix, "prefix", "x", "", "Set pairtree prefix")

}

func Run(args []string, writer io.Writer) error {
	var err error

	var rootCmd = &cobra.Command{
		Use:   "pt new -p [PT_ROOT]",
		Short: "pt new is a tool to create a Pairtree",
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
			if numArgs > 0 {
				fmt.Fprintln(writer, "There are too many arguments to ptcreate")
				Logger.Error("ptcreate should only have the pairtree root set and a possible prefix ",
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
		Logger.Error("Error setting command line", zap.Error(err))
		return err
	}

	// create the pairtree root directory if it does not exist
	if err = pairtree.CreatePairtree(ptRoot, prefix); err != nil {
		return err
	}

	return nil
}
