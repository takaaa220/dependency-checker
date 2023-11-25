package analyzer

import (
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

func NewDependencyCheckAnalyzer(pwd string, rules []Rule) *analysis.Analyzer {
	rulesByFilePath := make(map[string][]Rule)

	for _, rule := range rules {
		for _, filePattern := range rule.Files {
			allFiles, err := filepath.Glob(path.Join(pwd, filePattern))
			if err != nil {
				panic(err)
			}

			for _, file := range allFiles {
				rulesByFilePath[file] = append(rulesByFilePath[file], rule)
			}
		}
	}

	dependencyChecker := dependencyChecker{
		rulesByFilePath: rulesByFilePath,
	}

	return &analysis.Analyzer{
		Name: "dependency_check",
		Doc:  "dependency_check is a static analysis tool to check",
		Run:  dependencyChecker.run,
	}
}

type Rule struct {
	Files []string
	Allow []string
	Deny  []string
}

func (r Rule) disallow(importPath string) bool {
	if len(r.Allow) > 0 {
		if !slices.ContainsFunc(r.Allow, func(p string) bool {
			return strings.Contains(importPath, p)
		}) {
			return true
		}
	}
	if len(r.Deny) > 0 {
		if slices.ContainsFunc(r.Deny, func(p string) bool {
			return strings.Contains(importPath, p)
		}) {
			return true
		}
	}

	return false
}

type dependencyChecker struct {
	rulesByFilePath map[string][]Rule
}

func (d dependencyChecker) run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		pos := pass.Fset.Position(file.Pos())
		filePath := pos.Filename

		rules, found := d.rulesByFilePath[filePath]
		if !found {
			continue
		}

		for _, importSpec := range file.Imports {
			for _, rule := range rules {
				// remove double quote
				importPath := importSpec.Path.Value[1 : len(importSpec.Path.Value)-1]

				if rule.disallow(importPath) {
					pass.Report(analysis.Diagnostic{
						Pos:            importSpec.Pos(),
						End:            importSpec.End(),
						Category:       "dependency_check",
						Message:        fmt.Sprintf("import %s is not allowed", importPath),
						SuggestedFixes: nil,
					})
				}
			}
		}
	}

	return nil, nil
}
