package costeval

import "github.com/qw4990/tidb-cost-calibrator/utils"

func genTPCHQueries(n int) utils.Queries {
	var qs utils.Queries
	//qs = append(qs, genTPCHScan(n)...)
	//qs = append(qs, genTPCHAgg(n)...)
	//qs = append(qs, genTPCHJoin(n)...)
	//qs = append(qs, genTPCHSort(n)...)
	return qs
}

func genTPCHScan(n int) utils.Queries {
	return nil
}

func genTPCHAgg(n int) utils.Queries {
	return nil
}

func genTPCHJoin(n int) utils.Queries {
	return nil
}

func genTPCHSort(n int) utils.Queries {
	return nil
}
