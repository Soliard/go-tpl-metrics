package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"

	// Стандартные анализаторы (часть набора vet) из x/tools.
	"golang.org/x/tools/go/analysis/passes/assign"          // подозрительные присваивания
	"golang.org/x/tools/go/analysis/passes/atomic"          // корректное использование sync/atomic
	"golang.org/x/tools/go/analysis/passes/bools"           // упрощение булевых выражений
	"golang.org/x/tools/go/analysis/passes/buildtag"        // корректность build tags
	"golang.org/x/tools/go/analysis/passes/cgocall"         // предупреждения по cgo
	"golang.org/x/tools/go/analysis/passes/composite"       // составные литералы
	"golang.org/x/tools/go/analysis/passes/copylock"        // копирование структур с мьютексами
	"golang.org/x/tools/go/analysis/passes/deepequalerrors" // использование deep equal с ошибками
	"golang.org/x/tools/go/analysis/passes/errorsas"        // проверка правильности использования errors.As
	"golang.org/x/tools/go/analysis/passes/httpresponse"    // проверка обработки HTTP ответов
	"golang.org/x/tools/go/analysis/passes/ifaceassert"     // проверка утверждений типов интерфейсов
	"golang.org/x/tools/go/analysis/passes/loopclosure"     // захват переменных в замыканиях циклов
	"golang.org/x/tools/go/analysis/passes/lostcancel"      // потеря контекста отмены
	"golang.org/x/tools/go/analysis/passes/nilfunc"         // сравнение функций с nil
	"golang.org/x/tools/go/analysis/passes/printf"          // проверка форматированных выводов
	"golang.org/x/tools/go/analysis/passes/shadow"          // перекрытие переменных
	"golang.org/x/tools/go/analysis/passes/shift"           // проверка сдвигов битов
	"golang.org/x/tools/go/analysis/passes/stdmethods"      // стандартные методы (String, Error и др.)
	"golang.org/x/tools/go/analysis/passes/structtag"       // проверка тегов структур
	"golang.org/x/tools/go/analysis/passes/tests"           // распространенные ошибки в тестах
	"golang.org/x/tools/go/analysis/passes/unreachable"     // недостижимый код
	"golang.org/x/tools/go/analysis/passes/unusedresult"    // неиспользованные результаты функций

	// Staticcheck: SA (correctness) + другие классы (simple, stylecheck).
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	// Публичные анализаторы.
	"github.com/gostaticanalysis/nilerr"
	"github.com/tdakkota/asciicheck"
)

// Документация (godoc):
// Package main предоставляет мультичекер статического анализа для проекта.
// Как запустить:
//   go run ./cmd/staticlint ./...
// или собрать и затем:
//   staticlint ./...
//
// Состав анализаторов:
// - Стандартные из golang.org/x/tools/go/analysis/passes (vet-подобные проверки).
// - Все анализаторы класса SA из staticcheck (Correctness).
// - Дополнительно некоторые из других классов staticcheck:
//   * simple (серия S1000...) — упрощения;
//   * stylecheck (серия ST1000...) — стиль кода.
// - Два публичных анализатора: asciicheck (ASCII-символы в идентификаторах) и nilerr (ошибки, теряющиеся при возврате nil).
// - Собственный анализатор: запрет прямого os.Exit в функции main пакета main.
//
// Назначение ключевых анализаторов (кратко):
// - SA* (staticcheck): указывают на потенциальные ошибки в логике (correctness).
// - S1* (simple): предлагают упрощения.
// - ST* (stylecheck): стиль, именование, комментарии.
// - nilerr: находит возвраты nil при наличии ошибки.
// - asciicheck: подчёркивает не-ASCII символы в именах.
// - noOsExitInMain: запрещает прямой вызов os.Exit внутри main.

func main() {
	var list []*analysis.Analyzer

	// Стандартные анализаторы.
	list = append(list,
		assign.Analyzer, atomic.Analyzer, bools.Analyzer, buildtag.Analyzer,
		cgocall.Analyzer, composite.Analyzer, copylock.Analyzer, deepequalerrors.Analyzer,
		errorsas.Analyzer, httpresponse.Analyzer, ifaceassert.Analyzer, loopclosure.Analyzer,
		lostcancel.Analyzer, nilfunc.Analyzer, printf.Analyzer, shadow.Analyzer,
		shift.Analyzer, stdmethods.Analyzer, structtag.Analyzer, tests.Analyzer,
		unreachable.Analyzer, unusedresult.Analyzer,
	)

	// Все SA-анализаторы.
	for _, a := range staticcheck.Analyzers {
		list = append(list, a.Analyzer)
	}

	// Другие классы: simple + stylecheck.
	for _, a := range simple.Analyzers {
		list = append(list, a.Analyzer)
	}
	for _, a := range stylecheck.Analyzers {
		list = append(list, a.Analyzer)
	}

	// Публичные анализаторы.
	list = append(list, asciicheck.NewAnalyzer())
	list = append(list, nilerr.Analyzer)

	// Наш пользовательский анализатор.
	list = append(list, analyzerNoOsExit)

	multichecker.Main(list...)
}
