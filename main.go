package main

import (
	"fmt"

	"github.com/qw4990/tidb-cost-calibrator/costeval"
	"github.com/qw4990/tidb-cost-calibrator/regression"
	"github.com/qw4990/tidb-cost-calibrator/utils"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tidb cost calibrator",
		Short: "tidb cost calibrator",
	}

	costEvalOption = utils.Option{
		Addr:     "",
		Port:     0,
		User:     "root",
		Password: "",
		Label:    "",
	}
	dbName      string
	ceType      string
	minioOption = utils.MinioOption{}
)

func newCostEvalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost-eval",
		Short: "Cost Model Evaluation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(dbName) == 0 {
				dbName = ceType
			}
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
	costEvalCmd := newCostEvalCmd()
	costEvalCmd.PersistentFlags().StringVar(&costEvalOption.Addr, "host", "127.0.0.1", "Connect to host")
	costEvalCmd.PersistentFlags().IntVar(&costEvalOption.Port, "port", 4000, "Port number to use for connection")
	costEvalCmd.PersistentFlags().StringVar(&dbName, "dbname", "", "database name")
	costEvalCmd.PersistentFlags().StringVar(&ceType, "ce-type", "tpch_clustered", "Cost Eval Type")
	costEvalCmd.PersistentFlags().StringVar(&minioOption.Endpoint, "minio-endpoint", "", "minio endpoint")
	costEvalCmd.PersistentFlags().StringVar(&minioOption.ID, "minio-id", "", "minio id")
	costEvalCmd.PersistentFlags().StringVar(&minioOption.Secret, "minio-secret", "", "minio secret")
	rootCmd.AddCommand(costEvalCmd)
	rootCmd.AddCommand(newCostCaliCmd())
	rootCmd.AddCommand(newCostRegressionCmd())
	rootCmd.AddCommand(newRegressionDetectCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
