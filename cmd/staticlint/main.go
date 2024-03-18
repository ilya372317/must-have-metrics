// Multichecker with all SE linters from staticheck library and all main linters from
// golang.org/x/tools/go/analysis/passes package.
//
// Usage:
//
// In root directory of the project run command 'make my-lint'.
// Command will create directory my-lint with result.txt file.
// In this file you can see problems with project code and places where they can be found.
//
// You also can clear lint results. Just run 'make my-lint-clear' and command will remove folder my-lint
package main

import (
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/pgk/analysis/mainexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	staticConfig := config.NewStaticConfig()
	staticChecks := []*analysis.Analyzer{
		// Analyzer that detects if there is only one variable in append.
		appends.Analyzer,
		// Analyzer that reports mismatches between assembly files and Go declarations.
		asmdecl.Analyzer,
		// Analyzer that detects useless assignments.
		assign.Analyzer,
		// Analyzer that checks for common mistakes using the sync/atomic package.
		atomic.Analyzer,
		// Analyzer that checks for non-64-bit-aligned arguments to sync/atomic functions.
		atomicalign.Analyzer,
		// Analyzer that detects common mistakes involving boolean operators.
		bools.Analyzer,
		// Analyzer that constructs the SSA representation of an error-free
		// package and returns the set of all functions within it.
		buildssa.Analyzer,
		// Analyzer that checks build tags.
		buildtag.Analyzer,
		// Analyzer that detects some violations of the cgo pointer passing rules.
		cgocall.Analyzer,
		// Analyzer that checks for unkeyed composite literals.
		composite.Analyzer,
		// Analyzer that checks for locks erroneously passed by value.
		copylock.Analyzer,
		// Analyzer that provides a syntactic control-flow graph (CFG) for the body of a function.
		ctrlflow.Analyzer,
		// Analyzer that checks for the use of reflect.DeepEqual with error values.
		deepequalerrors.Analyzer,
		// Analyzer that checks for common mistakes in defer statements.
		defers.Analyzer,
		// Analyzer that checks known Go toolchain directives.
		directive.Analyzer,
		// Analyzer that checks that the second argument to errors.As is a pointer to a type implementing error.
		errorsas.Analyzer,
		// Analyzer that detects structs that would use less memory if their fields were sorted.
		fieldalignment.Analyzer,
		// Analyzer that serves as a trivial example and test of the Analysis API.
		findcall.Analyzer,
		// Analyzer that reports assembly code that clobbers the frame pointer before saving it.
		framepointer.Analyzer,
		// Analyzer that checks for mistakes using HTTP responses.
		httpresponse.Analyzer,
		// Analyzer that flags impossible interface-interface type assertions.
		ifaceassert.Analyzer,
		// Analyzer that checks for references to enclosing loop variables from within nested functions.
		loopclosure.Analyzer,
		// Analyzer that checks for failure to call a context cancellation function.
		lostcancel.Analyzer,
		// Analyzer that checks for useless comparisons against nil.
		nilfunc.Analyzer,
		// nilness inspects the control-flow graph of an SSA function and reports errors
		// such as nil pointer dereferences and degenerate nil pointer comparisons.
		nilness.Analyzer,
		// Analyzer that checks consistency of Printf format strings and arguments.
		printf.Analyzer,
		// Analyzer that checks for accidentally using == or reflect.DeepEqual to compare reflect.Value values.
		reflectvaluecompare.Analyzer,
		// Analyzer that checks for shifts that exceed the width of an integer.
		shift.Analyzer,
		// Analyzer that detects misuse of unbuffered signal as argument to signal.Notify.
		sigchanyzer.Analyzer,
		// Analyzer that checks for mismatched key-value pairs in log/slog calls.
		slog.Analyzer,
		// Analyzer that checks for calls to sort.Slice that do not use a slice type as first argument.
		sortslice.Analyzer,
		// Analyzer that checks for misspellings in the signatures of methods similar to well-known interfaces.
		stdmethods.Analyzer,
		// Analyzer that flags type conversions from integers to strings.
		stringintconv.Analyzer,
		// Analyzer that checks struct field tags are well-formed.
		structtag.Analyzer,
		// Analyzerfor detecting calls to Fatal from a test goroutine.
		testinggoroutine.Analyzer,
		// Analyzer that checks for common mistaken usages of tests and examples.
		tests.Analyzer,
		// Analyzer that checks for the use of time.Format or time.Parse calls with a bad format.
		timeformat.Analyzer,
		// Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions.
		unmarshal.Analyzer,
		// Analyzer that checks for unreachable code.
		unreachable.Analyzer,
		// Analyzer that checks for invalid conversions of uintptr to unsafe.Pointer.
		unsafeptr.Analyzer,
		// Analyzer that checks for unused results of calls to certain pure functions.
		unusedresult.Analyzer,
		// Analyzer checks for unused writes to the elements of a struct or array object.
		unusedwrite.Analyzer,
		// Analyzer that checks for usage of generic features added in Go 1.18.
		usesgenerics.Analyzer,
		// Analyzer that checks for main fucnction not calling os.Exit().
		mainexit.Analyzer,
	}

	// SE analyzers from staticheck library
	for _, check := range staticcheck.Analyzers {
		if _, ok := staticConfig.RuleSet[check.Analyzer.Name]; ok {
			staticChecks = append(staticChecks, check.Analyzer)
		}
	}

	multichecker.Main(staticChecks...)
}
