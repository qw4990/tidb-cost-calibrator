package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	templateDir := "/Users/zhangyuanjia/Workspace/go/src/github.com/pingCAP/ossinsight/api/queries"
	outputDir := "/Users/zhangyuanjia/Workspace/go/src/github.com/qw4990/tidb-cost-calibrator/ossinsight/bench_query"
	err := filepath.Walk(templateDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}
		tmp := strings.Split(path, "/")
		outputPath := filepath.Join(outputDir, tmp[len(tmp)-2]+".sql")

		data, err := os.ReadFile(path)
		fmt.Println(path, err)
		if err != nil {
			return err
		}
		return os.WriteFile(outputPath, data, 0666)
	})
	fmt.Println(err)
}
