package lint_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/takaaa220/dependency-checker/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test_Integration(t *testing.T) {
	// このファイルが配置されているディレクトリに存在するディレクトリをすべて取得する
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	dirs := []string{}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		dirs = append(dirs, file.Name())
	}

	for _, testName := range dirs {
		t.Run(testName, func(t *testing.T) {
			file, err := os.Open(filepath.Join(testName, "expect.json"))
			if err != nil {
				t.Errorf("failed to open expect.json: %v", err)
				return
			}

			var expect []Expect
			if err := json.NewDecoder(file).Decode(&expect); err != nil {
				t.Errorf("failed to decode expect.json: %v", err)
				return
			}

			pwd := os.Getenv("PWD")
			testdata := filepath.Join(pwd, testName, "testdata")

			res := analysistest.Run(t, testdata, analyzer.DependencyCheckAnalyzer, "./...")

			fmt.Println(res)
		})
	}
}

type Expect struct {
	Error string `json:error`
}
