package errorgen

import (
	"os"
	"testing"
)

func TestGenerateFromCSV(t *testing.T) {
	type args struct {
		templateFileDir   string
		templateName      string
		fileLocation      string
		outputDestination string
		outputFile        string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "success",
			args: args{
				fileLocation:      "./../../../../storages/errors-map.csv",
				templateFileDir:   "/error_map.tmpl",
				templateName:      "error_map.tmpl",
				outputDestination: "./",
				outputFile:        "error_map.go.go",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GenerateErrorMapFromCSV(tt.args.templateFileDir, tt.args.templateName, tt.args.fileLocation, tt.args.outputDestination, tt.args.outputFile)
		})
	}
	os.Remove("error_map.go.go")
}
