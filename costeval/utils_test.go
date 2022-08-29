package costeval

import (
	"fmt"
	"testing"
)

func TestGen4Template(t *testing.T) {
	for _, q := range gen4Templates([]template{
		{
			`select * from t where #`,
			[]tempitem{{"t", "a", 1, 100}},
			"TableScan",
		},
		{
			`select * from t where # and #`,
			[]tempitem{
				{"t", "a", 1, 100},
				{"t", "b", 100, 200},
			},
			"TableScan2",
		},
		{
			`select * from t t1, t t2 where t1.a=t2=a and # and #`,
			[]tempitem{
				{"t1", "a", 1, 100},
				{"t2", "b", 100, 200},
			},
			"TableScan2",
		},
	}, 5) {
		fmt.Println(q.Label, q.SQL)
	}
}
