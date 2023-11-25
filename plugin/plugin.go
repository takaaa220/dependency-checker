package plugin

import (
	"fmt"

	"github.com/takaaa220/dependency-checker/analyzer"
	"golang.org/x/tools/go/analysis"
)

func New(conf any) ([]*analysis.Analyzer, error) {
	fmt.Printf("My configuration (%[1]T): %#[1]v\n", conf)

	return []*analysis.Analyzer{analyzer.DependencyCheckAnalyzer}, nil
}
