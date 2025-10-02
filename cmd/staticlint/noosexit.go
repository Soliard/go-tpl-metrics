package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// analyzerNoOsExit запрещает прямой вызов os.Exit в функции main пакета main.
//
// Мотивация: прямой os.Exit в main затрудняет тестирование, не даёт выполнить отложенные функции (defer),
// и обычно лучше возвращать код ошибки через возврат из main или централизованно обрабатывать ошибки.
//
// Как работает:
// 1) Проверяем, что анализируемый пакет — это пакет main.
// 2) Ищем функцию с именем "main" на верхнем уровне.
// 3) Обходим её тело и репортим любые вызовы вида os.Exit(...).
var analyzerNoOsExit = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "forbids direct calls to os.Exit inside main() of package main",
	Run:  runNoOsExit,
}

func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	// Разрешаем анализ только пакета main.
	if pass.Pkg == nil || pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, f := range pass.Files {
		// Ищем верхнеуровневую функцию main
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != "main" || fn.Body == nil {
				continue
			}
			// Обходим тело функции main
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Ищем os.Exit(...)
				switch fun := call.Fun.(type) {
				case *ast.SelectorExpr:
					// ожидаем форму: os.Exit
					pkgIdent, ok := fun.X.(*ast.Ident)
					if !ok {
						return true
					}
					if pkgIdent.Name == "os" && fun.Sel != nil && fun.Sel.Name == "Exit" {
						pass.Reportf(call.Lparen, "do not call os.Exit directly in main(); return error or handle it gracefully")
					}
				}
				return true
			})
		}
	}

	return nil, nil
}
