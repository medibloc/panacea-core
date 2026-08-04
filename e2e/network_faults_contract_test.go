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

	"github.com/stretchr/testify/require"
)

func TestNetworkFaultWebSocketReconnectStaticContract(t *testing.T) {
	t.Parallel()

	exercise := networkFaultFunctionSource(t, "network_faults_test.go", "networkFaultExerciseWebSocketReconnect")
	start := networkFaultFunctionSource(t, "network_faults_test.go", "startNetworkFaultWebSocketSession")
	validation := networkFaultFunctionSource(t, "network_faults_test.go", "validateNetworkFaultWebSocketReconnect")

	require.Equal(t, 2, strings.Count(exercise, "collectNetworkFaultWebSocketPhase("))
	for _, contract := range []string{
		"startNetworkFaultWebSocketSession(",
		"ApplyFullNodeRPCBoundaryFault(",
		"RestoreFullNodeRPCBoundaryFault(",
		"probeNetworkFaultTCPConnectionFailure(",
		"validateNetworkFaultWebSocketReconnect(",
		"FaultMissingBlockHeights",
		`network-faults/websocket-reconnect.json`,
		`"same_client"`,
	} {
		require.Contains(t, exercise, contract)
	}
	require.Less(
		t,
		strings.Index(exercise, "startNetworkFaultWebSocketSession("),
		strings.Index(exercise, "ApplyFullNodeRPCBoundaryFault("),
		"the subscribed client must exist before the induced RPC fault",
	)
	for _, subscription := range []string{"EventNewBlock", "EventTx"} {
		require.Contains(t, start, subscription)
	}
	for _, defect := range []string{
		"DuplicateBlockEvents",
		"DuplicateTransactionEvents",
		"FaultMissingBlockHeights",
		"MissingTransactionHashes",
		"UnexpectedTransactionHashes",
		"has no matching block event",
	} {
		require.Contains(t, validation, defect)
	}
}

func TestNetworkFaultOversizedGRPCStaticContract(t *testing.T) {
	t.Parallel()

	exercise := networkFaultFunctionSource(t, "network_faults_test.go", "networkFaultExerciseGRPCMessageBoundary")
	validation := networkFaultFunctionSource(t, "network_faults_test.go", "validateOversizedGRPCRejection")
	for _, contract := range []string{
		"GRPCMaxRecvBytes",
		"QueryAllBalancesRequest",
		"proto.Size(",
		"WaitForNodeHeight(",
		"WaitForFullNode(",
		"RequireSameHistoryAtHeight(",
		`network-faults/oversized-grpc-message.json`,
		"normal_query_after_rejection",
	} {
		require.Contains(t, exercise, contract)
	}
	for _, contract := range []string{"codes.ResourceExhausted", "message larger than max"} {
		require.Contains(t, validation, contract)
	}
}

func TestNetworkFaultFailureCategoryStaticContract(t *testing.T) {
	t.Parallel()

	live := networkFaultFunctionSource(t, "network_faults_test.go", "TestLocalDockerNetworkAndEndpointFaults")
	category := networkFaultFunctionSource(t, filepath.Join("internal", "harness", "network_fault_category.go"), "RecordNetworkFaultCategory")
	for _, contract := range []string{
		"SetupFailureCategory",
		"NetworkFaultCategoryEnvironmentPreflight",
		"NetworkFaultOutcomeFailed",
		"initial-chain-readiness",
		"failure_category_artifacts",
	} {
		require.Contains(t, live, contract)
	}
	for _, artifact := range []string{
		"environment/network-failure-categories.jsonl",
		"network-faults/failure-categories.jsonl",
	} {
		require.Contains(t, category, artifact)
	}
}

func TestNetworkFaultCleanupStaticContract(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("network_faults_test.go")
	require.NoError(t, err)
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "network_faults_test.go", contents, 0)
	require.NoError(t, err)

	var violations []string
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Cleanup" || len(call.Args) != 1 {
			return true
		}
		receiver, ok := selector.X.(*ast.Ident)
		if !ok || receiver.Name != "t" {
			return true
		}
		cleanup, ok := call.Args[0].(*ast.FuncLit)
		if !ok {
			violations = append(violations, fileSet.Position(call.Pos()).String()+": t.Cleanup must receive a function literal")
			return true
		}
		body := string(contents[fileSet.Position(cleanup.Body.Pos()).Offset:fileSet.Position(cleanup.Body.End()).Offset])
		if !strings.Contains(body, "recordNetworkFaultCleanup(") {
			violations = append(violations, fileSet.Position(call.Pos()).String()+": network fault t.Cleanup does not record its result")
		}
		return true
	})

	implementationFiles, err := filepath.Glob(filepath.Join("internal", "harness", "network_fault*.go"))
	require.NoError(t, err)
	implementationFiles = append(implementationFiles, "network_faults_test.go")
	for _, path := range implementationFiles {
		if strings.HasSuffix(path, "_test.go") && path != "network_faults_test.go" {
			continue
		}
		source, readErr := os.ReadFile(path)
		if readErr != nil {
			violations = append(violations, path+": "+readErr.Error())
			continue
		}
		parsed, parseErr := parser.ParseFile(fileSet, path, source, 0)
		if parseErr != nil {
			violations = append(violations, path+": "+parseErr.Error())
			continue
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
				return true
			}
			blank, ok := assignment.Lhs[0].(*ast.Ident)
			if !ok || blank.Name != "_" {
				return true
			}
			if _, isCall := assignment.Rhs[0].(*ast.CallExpr); isCall {
				violations = append(violations, fileSet.Position(assignment.Pos()).String()+": single call result is discarded")
			}
			return true
		})
	}

	sort.Strings(violations)
	require.Empty(t, violations, strings.Join(violations, "\n"))
}

func networkFaultFunctionSource(t *testing.T, path, functionName string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, contents, 0)
	require.NoError(t, err)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != functionName {
			continue
		}
		start := fileSet.Position(function.Pos()).Offset
		end := fileSet.Position(function.End()).Offset
		return string(contents[start:end])
	}
	t.Fatalf("function %s not found in %s", functionName, path)
	return ""
}
