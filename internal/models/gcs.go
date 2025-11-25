package models

import (
	"fmt"
	"path/filepath"
	"strings"
)

type CloudStoragePayload struct {
	Filename string
	Path     string
}

func (c CloudStoragePayload) GetFilePath() string {
	return fmt.Sprintf("%s/%s", c.Path, c.Filename)
}

func NewCloudStoragePayload(input string) CloudStoragePayload {
	input = filepath.Clean(input)

	// Extract the directory and filename.
	path := filepath.Dir(input)
	filename := filepath.Base(input)

	// handle the special case where input might be just a filename.
	if strings.TrimSpace(path) == "." {
		path = "" // If there's no path, just set it to an empty string.
	}

	return CloudStoragePayload{Filename: filename, Path: path}
}

type DirectoryName string

const (
	TrialBalanceDetailDir DirectoryName = "trial_balance_detail"
	SubLedgerDir          DirectoryName = "sub_ledger"
	TrialBalanceDir       DirectoryName = "trial_balances"
)
