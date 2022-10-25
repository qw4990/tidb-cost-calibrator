package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

type Query struct {
	SQL   string
	Label string
}

type Queries []Query

type Record struct {
	Cost    float64
	TimeMS  float64
	Label   string
	SQL     string
	Weights map[string]float64
}

func (r Record) Clone() Record {
	rr := r
	rr.Weights = make(map[string]float64)
	for k, v := range r.Weights {
		rr.Weights[k] = v
	}
	return rr
}

type Records []Record

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func SaveTo(f string, r interface{}) {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetIndent("", "  ")
	jsonEncoder.SetEscapeHTML(false)
	Must(jsonEncoder.Encode(r))
	Must(ioutil.WriteFile(f, bf.Bytes(), 0666))
}

func ReadFrom(f string, r interface{}) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, r); err != nil {
		return err
	}
	return nil
}

func MustReadOneLine(ins Instance, q string, ret ...interface{}) {
	rs := ins.MustQuery(q)
	rs.Next()
	defer rs.Close()
	if err := rs.Scan(ret...); err != nil {
		panic(err)
	}
}
