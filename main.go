package main

import (
	"fmt"
	"github.com/qw4990/tidb-cost-calibrator/costeval"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tidb cost calibrator",
		Short: "tidb cost calibrator",
	}
)

func newCostEvalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost-eval",
		Short: "Cost Model Evaluation",
		RunE: func(cmd *cobra.Command, args []string) error {
			costeval.CostEval()
			return nil
		},
	}
	return cmd
}

func init() {
	cobra.OnInitialize()
	rootCmd.AddCommand(newCostEvalCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
