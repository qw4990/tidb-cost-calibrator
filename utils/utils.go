package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

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
