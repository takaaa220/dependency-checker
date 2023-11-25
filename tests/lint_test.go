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

			rulesFile, err := os.Open(filepath.Join(testName, "rules.json"))
			if err != nil {
				t.Errorf("failed to open rules.json: %v", err)
				return
			}

			var rules []Rule
			if err := json.NewDecoder(rulesFile).Decode(&rules); err != nil {
				t.Errorf("failed to decode rules.json: %v", err)
				return
			}

			analyzerRules := make([]analyzer.Rule, len(rules))
			for i, rule := range rules {
				analyzerRules[i] = analyzer.Rule{
					Files: rule.Files,
					Allow: rule.Allow,
					Deny:  rule.Deny,
				}
			}

			pwd := os.Getenv("PWD")
			testdata := filepath.Join(pwd, testName, "testdata")

			res := analysistest.Run(nopeTesting{}, testdata, analyzer.NewDependencyCheckAnalyzer(testdata, analyzerRules), "./...")

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

type Rule struct {
	Files []string `json:files`
	Allow []string `json:allow`
	Deny  []string `json:deny`
}

type nopeTesting struct{}

func (nopeTesting) Errorf(string, ...interface{}) {}
