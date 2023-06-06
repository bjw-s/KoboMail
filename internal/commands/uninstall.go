// Package config implements all commands of KoboMail
package commands

import (
	"github.com/bjw-s/kobomail/internal/kobomail"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall KoboMail completely",
	Long:  "Uninstall KoboMail completely.",
	RunE: func(cmd *cobra.Command, args []string) error {
		kobomail.KoboMailConfig = conf
		zap.S().Debugw("Running with configuration",
			zap.Any("configuration", conf),
		)
		return nil
	},
}
