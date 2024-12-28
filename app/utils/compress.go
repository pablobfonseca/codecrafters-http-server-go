package utils

import (
	"bytes"
	"compress/gzip"
)

func CompressString(s string) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	defer gzWriter.Close()

	_, err := gzWriter.Write([]byte(s))
	if err != nil {
		return nil, err
	}

	err = gzWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
