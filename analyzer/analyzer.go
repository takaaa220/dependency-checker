package analyzer

import (
	"fmt"
	"path/filepath"
	"regexp"

	"golang.org/x/tools/go/analysis"
)

type dependencyChecker struct {
	pwd     string
	setting Setting
}

func NewDependencyCheckAnalyzer(pwd string, setting Setting) *analysis.Analyzer {
	dependencyChecker := dependencyChecker{
		pwd:     pwd,
		setting: setting,
	}

	return &analysis.Analyzer{
		Name: "dependency_check",
		Doc:  "dependency_check is a static analysis tool to check",
		Run:  dependencyChecker.run,
	}
}

func (c dependencyChecker) findDenyRules(filePath string) []Rule {
	rules := []Rule{}

	for _, rule := range c.setting.Deny {
		relativeFilePath, err := filepath.Rel(c.pwd, filePath)
		if err != nil {
			panic(err)
		}

		matched := rule.matchFilePath(relativeFilePath)
		if matched {
			rules = append(rules, rule)
		}
	}

	return rules
}

type Setting struct {
	Deny []Rule
}

type Rule struct {
	From    string // file relative path from module root (regexp)
	To      string // import path (regexp)
	Message string // error message
}

func (r Rule) matchFilePath(relativeFilePath string) bool {
	matched, err := regexp.MatchString(r.From, relativeFilePath)
	if err != nil {
		panic(err)
	}

	return matched
}

// errorMessage, disallowed を返す
func (r Rule) matchImportPath(importPath string) bool {
	matched, err := regexp.MatchString(r.To, importPath)
	if err != nil {
		panic(err)
	}

	return matched
}

func (r Rule) errorMessage(importPath string) string {
	if r.Message != "" {
		return r.Message
	}

	return fmt.Sprintf("import %s is not allowed", importPath)
}

func (d dependencyChecker) run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		pos := pass.Fset.Position(file.Pos())
		filePath := pos.Filename

		rules := d.findDenyRules(filePath)
		if len(rules) == 0 {
			continue
		}

		for _, importSpec := range file.Imports {
			for _, rule := range rules {
				// remove double quote
				importPath := importSpec.Path.Value[1 : len(importSpec.Path.Value)-1]

				if matched := rule.matchImportPath(importPath); matched {
					pass.Report(analysis.Diagnostic{
						Pos:            importSpec.Pos(),
						End:            importSpec.End(),
						Category:       "dependency_check",
						Message:        rule.errorMessage(importPath),
						SuggestedFixes: nil,
					})
				}
			}
		}
	}

	return nil, nil
}
