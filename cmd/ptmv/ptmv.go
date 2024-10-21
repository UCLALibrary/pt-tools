package ptmv

/* ptmv is a tool that can move files in and out of the Pairtree structure */

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	"github.com/UCLALibrary/pt-tools/pkg/pairtree"
	"github.com/UCLALibrary/pt-tools/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	tar     bool
	ptRoot  string
	logFile string      = "logs.log"
	Logger  *zap.Logger = utils.Logger(logFile)
	src     string      = ""
	dest    string      = ""
)

func initFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&ptRoot, "pairtree", "p", "", "Set pairtree root directory")
	cmd.Flags().BoolVarP(&tar, "a", "a", false, "Produce a tar/gzipped output or unpack a tar/gzipped")
}

func Run(args []string, writer io.Writer) error {
	var err error

	var rootCmd = &cobra.Command{
		Use:   "ptmv [PT_ROOT] [ID] [/path/to/output/]",
		Short: "Ptmv is a tool that can move files in and out of the Pairtree structure",
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
			if numArgs < 2 {
				fmt.Fprintln(writer, "Please provide a source and destination for copied files")
				Logger.Error("There are not enough arguments to ptmv",
					zap.Error(error_msgs.Err9))

				return error_msgs.Err9
			}

			if numArgs == 2 {
				// Extract the src and dest
				src = args[numArgs-2]
				dest = args[numArgs-1]
			} else {
				fmt.Fprintln(writer, "Too many arguments were provided to ptmv")
				Logger.Error("Error parsing ptmv", zap.Error(error_msgs.Err8))

				return error_msgs.Err8
			}

			Logger.Info("Pairtree root is", zap.String("PAIRTREE_ROOT", ptRoot))

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

	srcIsPairtree := false
	// Determine if the src or dest is the pairtree
	if strings.HasPrefix(src, prefix) {
		if src, err = pairtree.CreatePP(src, ptRoot, prefix); err != nil {
			Logger.Error("Error creating pairpath", zap.Error(err))
			return err
		}
		src = filepath.Join(src)
		srcIsPairtree = true
	} else if strings.HasPrefix(dest, prefix) {
		if dest, err = pairtree.CreatePP(dest, ptRoot, prefix); err != nil {
			Logger.Error("Error creating pairpath", zap.Error(err))
			return err
		}
		if err = pairtree.CreateDirNotExist(dest); err != nil {
			return err
		}
		dest = filepath.Join(dest)
	} else {
		fmt.Fprintln(writer,
			"Neither the source or destination contains a prefix and is not a part of the pairtree")
		Logger.Error("Error verifying source and destination",
			zap.Error(error_msgs.Err10))
		return error_msgs.Err10
	}

	fmt.Printf("This is the src: %s \n", src)
	fmt.Printf("This is the dest: %s \n", dest)

	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("failed to remove %s: %w", dest, err)
	}

	if tar {
		if srcIsPairtree {
			if err = pairtree.TarGz(src, dest, prefix, true); err != nil {
				Logger.Error("Error compressing pairtree object", zap.Error(err))
				return err
			}
		} else {
			if err = pairtree.UnTarGz(src, dest); err != nil {
				Logger.Error("Error decompressing .tgz file", zap.Error(err))
				return err
			}
		}
	} else {

		finalDest, err := pairtree.CopyFileOrFolder(src, dest, true)

		if err != nil {
			Logger.Error("Error copying source to destination", zap.Error(err))
			return err
		} else {
			Logger.Info("Folder or file was successfully copied to",
				zap.String("destination of File or Folder", finalDest))
		}
	}

	if err := os.RemoveAll(src); err != nil {
		return fmt.Errorf("failed to remove %s: %w", src, err)
	}
	return nil
}
