package file

import (
	"archive/zip"
	"bytes"
	"time"
)

func CompressCSVToZip(filename string, csvData []byte) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	header := &zip.FileHeader{
		Name:     filename,
		Method:   zip.Deflate,
		Modified: time.Now(),
	}

	f, err := zipWriter.CreateHeader(header)
	if err != nil {
		return nil, err
	}

	_, err = f.Write(csvData)
	if err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
