package file

import (
	"context"
	"encoding/csv"
	"io"
)

func (c *ioFile) NewCSVWriter(w io.Writer) *csv.Writer {
	c.Writer = csv.NewWriter(w)
	return c.Writer
}

func (c *ioFile) CSVWriteHeader(ctx context.Context, header []string) (err error) {
	err = c.Writer.Write(header)
	return
}

func (c *ioFile) CSVWriteBody(ctx context.Context, body []string) (err error) {
	err = c.Writer.Write(body)
	return
}

func (c *ioFile) CSVWriteAll(ctx context.Context, all [][]string) (err error) {
	err = c.Writer.WriteAll(all)
	return
}

func (c *ioFile) CSVProcessWrite(ctx context.Context) (err error) {
	c.Writer.Flush()
	err = c.Writer.Error()
	return
}

func (c *ioFile) CSVReadAll(fs io.Reader) (records [][]string, err error) {
	records, err = csv.NewReader(fs).ReadAll()
	return
}
