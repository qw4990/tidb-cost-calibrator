package utils

type Query struct {
	SQL   string
	Label string
}

type Queries []Query

type Record struct {
	Cost   float64
	TimeMS float64
	Label  string
	SQL    string
}

type Records []Record
