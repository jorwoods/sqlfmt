package cmd

import (
	"fmt"
	"os"

	"github.com/jorwoods/sqlfmt/formatter"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sqlfmt [file]",
	Short: "Formats Snowflake SQL according to your style guide",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]
		data, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		result := formatter.FormatSQL(string(data))
		fmt.Println(result)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
