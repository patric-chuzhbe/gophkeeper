package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vault-cli",
	Short: "Secure CLI client for managing encrypted data",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
