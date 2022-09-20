package main

import (
	"fmt"

	"github.com/qw4990/tidb-cost-calibrator/costeval"
	"github.com/qw4990/tidb-cost-calibrator/regression"
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

func newCostCaliCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost-cali",
		Short: "Cost Model Calibration",
		RunE: func(cmd *cobra.Command, args []string) error {
			costeval.CostCalibrate()
			return nil
		},
	}
	return cmd
}

func newCostRegressionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost-reg",
		Short: "Cost Model Regression",
		RunE: func(cmd *cobra.Command, args []string) error {
			costeval.CostRegression()
			return nil
		},
	}
	return cmd
}

func newRegressionDetectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reg-detect",
		Short: "Regression Detect",
		RunE: func(cmd *cobra.Command, args []string) error {
			regression.RegDetect()
			return nil
		},
	}
	return cmd
}

func init() {
	cobra.OnInitialize()
	rootCmd.AddCommand(newCostEvalCmd())
	rootCmd.AddCommand(newCostCaliCmd())
	rootCmd.AddCommand(newCostRegressionCmd())
	rootCmd.AddCommand(newRegressionDetectCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
