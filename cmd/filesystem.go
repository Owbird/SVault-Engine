package cmd

import (
	"github.com/Owbird/SVault-Engine/pkg/filesystem"
	"github.com/spf13/cobra"
)

// filesystemCmd represents the filesystem command
var filesystemCmd = &cobra.Command{
	Use:   "filesystem",
	Short: "The SVault filesystem",
	Long:  `The SVault filesystem`,
}

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount the SVault VFS",
	Long:  `Mount the SVault VFS`,
	Run: func(cmd *cobra.Command, args []string) {
		filesystem.Mount()
	},
}

func init() {
	rootCmd.AddCommand(filesystemCmd)

	filesystemCmd.AddCommand(mountCmd)
}
