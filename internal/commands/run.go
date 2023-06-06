// Package config implements all commands of KoboMail
package commands

import (
	"github.com/bjw-s/kobomail/internal/kobomail"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run KoboMail processing",
	Long:  "Run KoboMail processing.",
	RunE: func(cmd *cobra.Command, args []string) error {
		kobomail.KoboMailConfig = conf
		zap.S().Debugw("Running with configuration",
			zap.Any("configuration", conf),
		)
		kobomail.PreparePrerequisites()
		kobomail.Run()
		return nil
	},
}
