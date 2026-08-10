package e2e_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var panicGuardConstructors = map[string]struct{}{
	"Start":            {},
	"StartFromExport":  {},
	"StartIBCTopology": {},
}

// TestHarnessConstructorsInstallPanicRecorderImmediately prevents a Go 1.23
// testing edge case from silently returning: during panic unwinding,
// testing.T.Cleanup can still observe t.Failed() == false. Every live harness
// owner must therefore install its direct panic recorder immediately after the
// constructor result has been validated.
func TestHarnessConstructorsInstallPanicRecorderImmediately(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read e2e source directory: %v", err)
	}

	fileSet := token.NewFileSet()
	var violations []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		path := filepath.Clean(entry.Name())
		file, parseErr := parser.ParseFile(fileSet, path, nil, 0)
		if parseErr != nil {
			violations = append(violations, path+": parse source: "+parseErr.Error())
			continue
		}

		violations = append(violations, panicGuardASTViolations(fileSet, file)...)
	}

	if len(violations) == 0 {
		return
	}
	sort.Strings(violations)
	t.Fatalf("live harness panic-guard contract violations:\n%s", strings.Join(violations, "\n"))
}

func TestPanicGuardASTContractRejectsMissingOrLateDefer(t *testing.T) {
	t.Parallel()

	for _, constructor := range []string{"Start", "StartFromExport", "StartIBCTopology"} {
		constructor := constructor
		t.Run(constructor, func(t *testing.T) {
			t.Parallel()
			validSource := `package e2e_test
func live() {
	owner, err := harness.` + constructor + `()
	require.NoError(t, err)
	defer owner.RecordTestPanic()
}`
			fileSet := token.NewFileSet()
			file, err := parser.ParseFile(fileSet, "valid_"+constructor+".go", validSource, 0)
			if err != nil {
				t.Fatalf("parse valid fixture: %v", err)
			}
			if violations := panicGuardASTViolations(fileSet, file); len(violations) != 0 {
				t.Fatalf("valid fixture violations: %v", violations)
			}

			lateSource := `package e2e_test
func live() {
	owner, err := harness.` + constructor + `()
	require.NoError(t, err)
	use(owner)
	defer owner.RecordTestPanic()
}`
			fileSet = token.NewFileSet()
			file, err = parser.ParseFile(fileSet, "late_"+constructor+".go", lateSource, 0)
			if err != nil {
				t.Fatalf("parse late fixture: %v", err)
			}
			violations := panicGuardASTViolations(fileSet, file)
			if len(violations) != 1 || !strings.Contains(violations[0], "followed immediately") {
				t.Fatalf("late defer violations = %v, want one immediate-defer violation", violations)
			}
		})
	}
}

func panicGuardASTViolations(fileSet *token.FileSet, file *ast.File) []string {
	constructorCalls := make(map[token.Pos]string)
	validatedCalls := make(map[token.Pos]struct{})
	var violations []string
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if constructor := panicGuardConstructorName(call); constructor != "" {
			constructorCalls[call.Pos()] = constructor
		}
		return true
	})

	ast.Inspect(file, func(node ast.Node) bool {
		block, ok := node.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for statementIndex, statement := range block.List {
			call, constructor, receiver, errorResult, ok := panicGuardConstructorAssignment(statement)
			if !ok {
				continue
			}
			validatedCalls[call.Pos()] = struct{}{}
			position := fileSet.Position(call.Pos())
			if statementIndex+2 >= len(block.List) {
				violations = append(violations, panicGuardViolation(position, constructor, receiver, "missing result assertion followed by direct defer"))
				continue
			}
			normalStart := panicGuardRequireAssertion(block.List[statementIndex+1], "NoError", errorResult)
			expectedFailure := panicGuardRequireAssertion(block.List[statementIndex+1], "NotNil", receiver)
			if !normalStart && !expectedFailure {
				violations = append(violations, panicGuardViolation(position, constructor, receiver, "next statement must be require.NoError(error), or require.NotNil(owner) for an expected startup failure"))
				continue
			}
			if !panicGuardDirectDefer(block.List[statementIndex+2], receiver) {
				violations = append(violations, panicGuardViolation(position, constructor, receiver, "result assertion must be followed immediately by defer "+receiver+".RecordTestPanic()"))
				continue
			}
			if expectedFailure && (statementIndex+3 >= len(block.List) ||
				!panicGuardRequireAssertion(block.List[statementIndex+3], "Error", errorResult)) {
				violations = append(violations, panicGuardViolation(position, constructor, receiver, "expected-failure owner guard and defer must be followed immediately by require.Error(error)"))
			}
		}
		return true
	})

	for position, constructor := range constructorCalls {
		if _, ok := validatedCalls[position]; ok {
			continue
		}
		violations = append(violations, fileSet.Position(position).String()+": harness."+constructor+" must be a direct assignment guarded by RecordTestPanic")
	}
	return violations
}

func TestPanicGuardRequireAssertionMatchesMethodAndConstructorResult(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		source     string
		method     string
		identifier string
		want       bool
	}{
		"normal constructor error":   {source: "require.NoError(t, err)", method: "NoError", identifier: "err", want: true},
		"expected failure owner":     {source: "require.NotNil(t, network)", method: "NotNil", identifier: "network", want: true},
		"expected constructor error": {source: "require.Error(t, err)", method: "Error", identifier: "err", want: true},
		"different error":            {source: "require.NoError(t, otherErr)", method: "NoError", identifier: "err", want: false},
		"different package":          {source: "assert.NoError(t, err)", method: "NoError", identifier: "err", want: false},
		"different method":           {source: "require.NotNil(t, network)", method: "NoError", identifier: "network", want: false},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			expression, err := parser.ParseExpr(test.source)
			if err != nil {
				t.Fatalf("parse assertion: %v", err)
			}
			got := panicGuardRequireAssertion(
				&ast.ExprStmt{X: expression},
				test.method,
				test.identifier,
			)
			if got != test.want {
				t.Fatalf("panicGuardRequireAssertion(%s) = %t, want %t", test.source, got, test.want)
			}
		})
	}
}

func panicGuardConstructorName(call *ast.CallExpr) string {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	packageName, ok := selector.X.(*ast.Ident)
	if !ok || packageName.Name != "harness" {
		return ""
	}
	if _, ok := panicGuardConstructors[selector.Sel.Name]; !ok {
		return ""
	}
	return selector.Sel.Name
}

func panicGuardConstructorAssignment(statement ast.Stmt) (
	call *ast.CallExpr,
	constructor string,
	receiver string,
	errorResult string,
	ok bool,
) {
	assignment, ok := statement.(*ast.AssignStmt)
	if !ok || len(assignment.Rhs) != 1 || len(assignment.Lhs) == 0 {
		return nil, "", "", "", false
	}
	call, ok = assignment.Rhs[0].(*ast.CallExpr)
	if !ok {
		return nil, "", "", "", false
	}
	constructor = panicGuardConstructorName(call)
	if constructor == "" {
		return nil, "", "", "", false
	}
	receiverIdentifier, ok := assignment.Lhs[0].(*ast.Ident)
	if !ok || receiverIdentifier.Name == "_" {
		return nil, "", "", "", false
	}
	receiver = receiverIdentifier.Name
	if len(assignment.Lhs) > 1 {
		if errorIdentifier, isIdentifier := assignment.Lhs[1].(*ast.Ident); isIdentifier {
			errorResult = errorIdentifier.Name
		}
	}
	return call, constructor, receiver, errorResult, true
}

func panicGuardRequireAssertion(statement ast.Stmt, method, identifier string) bool {
	expression, ok := statement.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := expression.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	if !ok || packageName.Name != "require" {
		return false
	}
	return selector.Sel.Name == method && panicGuardCallHasIdentifier(call, identifier)
}

func panicGuardCallHasIdentifier(call *ast.CallExpr, name string) bool {
	if name == "" {
		return false
	}
	for _, argument := range call.Args {
		identifier, ok := argument.(*ast.Ident)
		if ok && identifier.Name == name {
			return true
		}
	}
	return false
}

func panicGuardDirectDefer(statement ast.Stmt, receiver string) bool {
	deferStatement, ok := statement.(*ast.DeferStmt)
	if !ok || len(deferStatement.Call.Args) != 0 {
		return false
	}
	selector, ok := deferStatement.Call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "RecordTestPanic" {
		return false
	}
	identifier, ok := selector.X.(*ast.Ident)
	return ok && identifier.Name == receiver
}

func panicGuardViolation(position token.Position, constructor, receiver, problem string) string {
	return position.String() + ": harness." + constructor + " result " + receiver + ": " + problem
}
