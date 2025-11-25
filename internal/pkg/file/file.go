package file

import (
	"context"
	"encoding/csv"
	"io"
	"os"
)

type IOFile interface {
	OpenFile(path string) (dst *os.File, err error)
	CreateFile(path string) (dst *os.File, err error)
	RemoveFile(path string) (err error)
	CopyFile(dst io.Writer, src io.Reader) (err error)
	ReadAll(src io.Reader) ([]byte, error)

	NewCSVWriter(w io.Writer) *csv.Writer
	CSVWriteHeader(ctx context.Context, header []string) (err error)
	CSVWriteBody(ctx context.Context, body []string) (err error)
	CSVWriteAll(ctx context.Context, all [][]string) (err error)
	CSVProcessWrite(ctx context.Context) (err error)
	CSVReadAll(fs io.Reader) (records [][]string, err error)
}

type ioFile struct {
	Writer *csv.Writer
}

func New() IOFile {
	return &ioFile{}
}

func (c *ioFile) OpenFile(path string) (dst *os.File, err error) {
	dst, err = os.Open(path)
	return
}

func (c *ioFile) CreateFile(path string) (dst *os.File, err error) {
	dst, err = os.Create(path)
	return
}

func (c *ioFile) RemoveFile(path string) (err error) {
	err = os.Remove(path)
	return
}

func (c *ioFile) CopyFile(dst io.Writer, src io.Reader) (err error) {
	_, err = io.Copy(dst, src)
	return
}

func (c *ioFile) ReadAll(src io.Reader) ([]byte, error) {
	return io.ReadAll(src)
}
