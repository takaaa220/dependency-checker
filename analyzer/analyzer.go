package analyzer

import (
	"fmt"

	"golang.org/x/tools/go/analysis"
)

var DependencyCheckAnalyzer = &analysis.Analyzer{
	Name: "dependency_check",
	Doc:  "dependency_check is a static analysis tool to check",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		fmt.Println(file)

		for _, importSpec := range file.Imports {
			fmt.Println(importSpec)
		}
	}

	return nil, nil
}
