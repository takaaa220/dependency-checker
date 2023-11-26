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

func NewDependencyCheckAnalyzer(pwd string, setting Setting) (*analysis.Analyzer, error) {
	dependencyChecker := dependencyChecker{
		pwd:     pwd,
		setting: setting,
	}

	err := dependencyChecker.compile()
	if err != nil {
		return nil, err
	}

	return &analysis.Analyzer{
		Name: "dependency_check",
		Doc:  "dependency_check is a static analysis tool to check",
		Run:  dependencyChecker.run,
	}, nil
}

func (c dependencyChecker) compile() error {
	return c.setting.compile()
}

func (c dependencyChecker) findDenyRules(filePath string) []Rule {
	return c.findRules(filePath, c.setting.Deny)
}

func (c dependencyChecker) findAllowRules(filePath string) []Rule {
	return c.findRules(filePath, c.setting.Allow)
}

func (c dependencyChecker) findRules(filePath string, allRules []Rule) []Rule {
	rules := []Rule{}
	for _, rule := range allRules {
		matched := rule.matchFilePath(filePath)
		if matched {
			rules = append(rules, rule)
		}
	}

	return rules
}

type Setting struct {
	Deny  []Rule
	Allow []Rule
}

func (s Setting) compile() error {
	for i, rule := range s.Deny {
		fromRegexp, err := regexp.Compile(rule.From)
		if err != nil {
			return fmt.Errorf("failed to compile regexp from %s: %w", rule.From, err)
		}

		toRegexp, err := regexp.Compile(rule.To)
		if err != nil {
			return fmt.Errorf("failed to compile regexp to %s: %w", rule.To, err)
		}

		s.Deny[i].fromRegexp = fromRegexp
		s.Deny[i].toRegexp = toRegexp
	}

	for i, rule := range s.Allow {
		fromRegexp, err := regexp.Compile(rule.From)
		if err != nil {
			return fmt.Errorf("failed to compile regexp from %s: %w", rule.From, err)
		}

		toRegexp, err := regexp.Compile(rule.To)
		if err != nil {
			return fmt.Errorf("failed to compile regexp to %s: %w", rule.To, err)
		}

		s.Allow[i].fromRegexp = fromRegexp
		s.Allow[i].toRegexp = toRegexp
	}

	return nil
}

type Rule struct {
	From       string // file relative path from module root (regexp)
	To         string // import path (regexp)
	Message    string // error message
	fromRegexp *regexp.Regexp
	toRegexp   *regexp.Regexp
}

func (r Rule) matchFilePath(relativeFilePath string) bool {
	return r.fromRegexp.MatchString(relativeFilePath)
}

// errorMessage, disallowed を返す
func (r Rule) matchImportPath(importPath string) bool {
	return r.toRegexp.MatchString(importPath)
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

		relativeFilePath, err := filepath.Rel(d.pwd, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative file path (pwd: %s, filepath: %s): %w", d.pwd, filePath, err)
		}

		allowRules := d.findAllowRules(relativeFilePath)
		denyRules := d.findDenyRules(relativeFilePath)
		if len(allowRules) == 0 && len(denyRules) == 0 {
			continue
		}

		for _, importSpec := range file.Imports {
			// remove double quote
			importPath := importSpec.Path.Value[1 : len(importSpec.Path.Value)-1]

			for _, rule := range allowRules {
				if matched := rule.matchImportPath(importPath); !matched {
					pass.Report(analysis.Diagnostic{
						Pos:            importSpec.Pos(),
						End:            importSpec.End(),
						Category:       "dependency_check",
						Message:        rule.errorMessage(importPath),
						SuggestedFixes: nil,
					})
					break
				}
			}

			for _, rule := range denyRules {
				if matched := rule.matchImportPath(importPath); matched {
					pass.Report(analysis.Diagnostic{
						Pos:            importSpec.Pos(),
						End:            importSpec.End(),
						Category:       "dependency_check",
						Message:        rule.errorMessage(importPath),
						SuggestedFixes: nil,
					})
					break
				}
			}
		}
	}

	return nil, nil
}
