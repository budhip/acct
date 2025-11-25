package main

import (
	"bitbucket.org/Amartha/go-accounting/internal/pkg/codegen/errorgen"
)

var (
	fileLocation      = "./storages/errors-map.csv"
	templateFileDir   = "/internal/pkg/codegen/errorgen/error_map.tmpl"
	templateName      = "error_map.tmpl"
	outputDestination = "./internal/models/"
	outputFile        = "error_map.go"
)

func main() {
	errorgen.GenerateErrorMapFromCSV(
		templateFileDir,
		templateName,
		fileLocation,
		outputDestination,
		outputFile)
}
