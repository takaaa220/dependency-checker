package lint_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
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
			expectFile, err := os.Open(filepath.Join(testName, "expect.json"))
			if err != nil {
				t.Errorf("failed to open expect.json: %v", err)
				return
			}

			var expect []Result
			if err := json.NewDecoder(expectFile).Decode(&expect); err != nil {
				t.Errorf("failed to decode expect.json: %v", err)
				return
			}

			rulesFile, err := os.Open(filepath.Join(testName, "setting.json"))
			if err != nil {
				t.Errorf("failed to open rules.json: %v", err)
				return
			}

			var setting Setting
			if err := json.NewDecoder(rulesFile).Decode(&setting); err != nil {
				t.Errorf("failed to decode rules.json: %v", err)
				return
			}

			analyzerSetting := analyzer.Setting{
				Deny:  make([]analyzer.Rule, len(setting.Deny)),
				Allow: make([]analyzer.Rule, len(setting.Allow)),
			}

			for i, rule := range setting.Deny {
				analyzerSetting.Deny[i] = analyzer.Rule{
					From:    rule.From,
					To:      rule.To,
					Message: rule.Message,
				}
			}
			for i, rule := range setting.Allow {
				analyzerSetting.Allow[i] = analyzer.Rule{
					From:    rule.From,
					To:      rule.To,
					Message: rule.Message,
				}
			}

			pwd := os.Getenv("PWD")
			testdata := filepath.Join(pwd, testName, "testdata")

			a, err := analyzer.NewDependencyCheckAnalyzer(testdata, analyzerSetting)
			if err != nil {
				t.Errorf("failed to create analyzer: %v", err)
				return
			}

			res := analysistest.Run(nopeTesting{}, testdata, a, "./...")

			result := []Result{}
			for _, rep := range res {
				for _, diagnostic := range rep.Diagnostics {
					pos := rep.Pass.Fset.Position(diagnostic.Pos)
					end := rep.Pass.Fset.Position(diagnostic.End)

					result = append(result, Result{
						FileName: pos.Filename,
						Start:    pos.Offset,
						End:      end.Offset,
						Message:  diagnostic.Message,
					})
				}
			}

			sortOpt := cmp.Transformer("Sort", func(in []Result) []Result {
				out := append([]Result(nil), in...)
				sort.Slice(out, func(i, j int) bool {
					return out[i].FileName < out[j].FileName
				})
				return out
			})

			if diff := cmp.Diff(expect, result, sortOpt); diff != "" {
				t.Errorf("missmatch result: %v", diff)
			}
		})
	}
}

type Result struct {
	FileName string `json:fileName`
	Start    int    `json:start`
	End      int    `json:end`
	Message  string `json:error`
}

type Setting struct {
	Deny  []Rule `json:deny`
	Allow []Rule `json:allow`
}

type Rule struct {
	From    string `json:from`
	To      string `json:to`
	Message string `json:message`
}

type nopeTesting struct{}

func (nopeTesting) Errorf(string, ...interface{}) {}
