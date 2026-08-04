package harness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const legacyDIDMnemonicRedaction = "A random mnemonic was generated: [REDACTED]"

const legacyDIDPasswordFileName = ".panacea-e2e-did-password"

var legacyDIDMnemonicLine = regexp.MustCompile(`A random mnemonic was generated:[^\r\n]*`)

func legacyDIDCreateCommand(passwordFile string, txCommand []string) []string {
	command := []string{
		"sh", "-c",
		`set -eu
umask 077
password_file=$1
shift
if [ ! -s "$password_file" ]; then
	random_password=$(od -An -N24 -tx1 /dev/urandom | tr -d ' \n')
	[ -n "$random_password" ]
	printf 'panacea-e2e-%s\n' "$random_password" >"$password_file"
fi
tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/panacea-e2e-did.XXXXXX")
stderr_fifo="$tmp_dir/stderr"
filter_pid=
cleanup() {
	if [ -n "$filter_pid" ]; then
		kill "$filter_pid" 2>/dev/null || true
	fi
	rm -f "$stderr_fifo"
	rmdir "$tmp_dir" 2>/dev/null || true
}
trap cleanup 0
trap 'exit 129' 1
trap 'exit 130' 2
trap 'exit 143' 15
mkfifo "$stderr_fifo"
sed 's/A random mnemonic was generated: .*/A random mnemonic was generated: [REDACTED]/' <"$stderr_fifo" >&2 &
filter_pid=$!
IFS= read -r password <"$password_file"
[ -n "$password" ]
set +e
printf '%s\n%s\n' "$password" "$password" | "$@" 2>"$stderr_fifo"
command_status=$?
wait "$filter_pid"
filter_status=$?
filter_pid=
set -e
if [ "$command_status" -ne 0 ]; then
	exit "$command_status"
fi
exit "$filter_status"`,
		"did-create",
		passwordFile,
	}
	return append(command, txCommand...)
}

func legacyDIDAuthenticatedCommand(passwordFile string, txCommand []string) []string {
	command := []string{
		"sh", "-c",
		`set -eu
password_file=$1
shift
[ -s "$password_file" ]
IFS= read -r password <"$password_file"
[ -n "$password" ]
printf '%s\n' "$password" | "$@"`,
		"did-authenticated",
		passwordFile,
	}
	return append(command, txCommand...)
}

func sanitizeLegacyDIDOutput(output []byte) []byte {
	return legacyDIDMnemonicLine.ReplaceAll(output, []byte(legacyDIDMnemonicRedaction))
}

func sanitizeLegacyDIDError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(string(sanitizeLegacyDIDOutput([]byte(err.Error()))))
}

// BroadcastDIDCreateAndWaitTx runs the legacy DID CLI with two matching,
// disposable password lines. The old binary prints its generated DID mnemonic
// to stderr, so every output and error path is sanitized before it can enter an
// artifact or assertion message.
func (n *Network) BroadcastDIDCreateAndWaitTx(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	command ...string,
) (*TxResult, error) {
	return n.broadcastDIDAndWaitTx(
		ctx,
		step,
		node,
		keyName,
		"matching-disposable-password-lines",
		func(txCommand []string) []string {
			return legacyDIDCreateCommand(
				path.Join(node.HomeDir(), legacyDIDPasswordFileName),
				txCommand,
			)
		},
		command...,
	)
}

// BroadcastDIDAuthenticatedAndWaitTx runs update-did or deactivate-did with
// the disposable password created inside the node home by
// BroadcastDIDCreateAndWaitTx. The password and encrypted private key never
// cross the Docker boundary or enter an artifact.
func (n *Network) BroadcastDIDAuthenticatedAndWaitTx(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	command ...string,
) (*TxResult, error) {
	return n.broadcastDIDAndWaitTx(
		ctx,
		step,
		node,
		keyName,
		"node-local-disposable-password",
		func(txCommand []string) []string {
			return legacyDIDAuthenticatedCommand(
				path.Join(node.HomeDir(), legacyDIDPasswordFileName),
				txCommand,
			)
		},
		command...,
	)
}

func (n *Network) broadcastDIDAndWaitTx(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	stdinMode string,
	wrap func([]string) []string,
	command ...string,
) (*TxResult, error) {
	n.txMu.Lock()
	defer n.txMu.Unlock()

	if strings.TrimSpace(step) == "" {
		return nil, errors.New("transaction step is required")
	}
	if node == nil {
		return nil, errors.New("transaction node is required")
	}
	if strings.TrimSpace(keyName) == "" {
		return nil, errors.New("transaction key name is required")
	}
	if len(command) == 0 {
		return nil, errors.New("transaction command is required")
	}
	if wrap == nil {
		return nil, errors.New("DID transaction wrapper is required")
	}

	requestID := fmt.Sprintf("%s-%d", step, time.Now().UTC().UnixNano())
	if err := n.artifacts.appendJSONLine("tx/requests.jsonl", map[string]any{
		"request_id":  requestID,
		"recorded_at": time.Now().UTC(),
		"step":        step,
		"node":        node.Name(),
		"key_name":    keyName,
		"arguments":   append([]string(nil), command...),
		"stdin_mode":  stdinMode,
	}); err != nil {
		return nil, fmt.Errorf("record DID transaction request %s: %w", step, err)
	}

	txCommand := node.TxCommand(keyName, command...)
	shellCommand := wrap(txCommand)
	stdout, stderr, execErr := node.Exec(ctx, shellCommand, node.Chain.Config().Env)
	safeStdout := sanitizeLegacyDIDOutput(stdout)
	safeStderr := sanitizeLegacyDIDOutput(stderr)
	safeExecErr := sanitizeLegacyDIDError(execErr)
	if err := n.artifacts.appendJSONLine("tx/broadcast-results.jsonl", map[string]any{
		"request_id":  requestID,
		"recorded_at": time.Now().UTC(),
		"step":        step,
		"stdout":      jsonOrString(safeStdout),
		"stderr":      boundedString(safeStderr, txStderrMaxBytes),
		"exec_error":  errorString(safeExecErr),
	}); err != nil {
		return nil, fmt.Errorf("record DID transaction broadcast %s: %w", step, err)
	}
	if execErr != nil {
		return nil, fmt.Errorf("broadcast DID transaction %s: %w: %s", step, safeExecErr, boundedString(safeStderr, txStderrMaxBytes))
	}

	broadcast, err := parseTxResult(stdout)
	if err != nil {
		return nil, fmt.Errorf("decode DID CheckTx response for %s: %w", step, err)
	}
	if broadcast.Code != 0 {
		return &broadcast, fmt.Errorf(
			"DID CheckTx %s failed: codespace=%s code=%d raw_log=%s",
			step,
			broadcast.Codespace,
			broadcast.Code,
			broadcast.RawLog,
		)
	}
	committed, err := n.waitForCommittedTx(ctx, requestID, step, broadcast.TxHash)
	if err != nil {
		return committed, err
	}
	if committed.Code != 0 {
		return committed, fmt.Errorf(
			"DID FinalizeBlock %s failed: codespace=%s code=%d raw_log=%s",
			step,
			committed.Codespace,
			committed.Code,
			committed.RawLog,
		)
	}
	return committed, nil
}

// DIDVerificationMethodIDs returns only public verification-method identifiers
// from the CLI keystore. Encrypted key documents never cross the node boundary
// or enter artifacts.
func (n *Network) DIDVerificationMethodIDs(ctx context.Context, node *cosmos.ChainNode) ([]string, error) {
	if node == nil {
		return nil, errors.New("DID key node is required")
	}
	script := `set -eu
home=$1
set -- "$home"/did_keystore/UTC--*.json
[ -f "$1" ]
for key_file do
	address=$(sed -n 's/.*"address":"\([^"]*\)".*/\1/p' "$key_file")
	[ -n "$address" ]
	printf '%s\n' "$address"
done`
	stdout, stderr, err := node.Exec(
		ctx,
		[]string{"sh", "-c", script, "did-public-identifiers", node.HomeDir()},
		node.Chain.Config().Env,
	)
	if err != nil {
		return nil, fmt.Errorf("read public DID key identifiers: %w: %s", err, boundedString(stderr, txStderrMaxBytes))
	}
	lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	identifiers := make([]string, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		identifier := strings.TrimSpace(line)
		if identifier == "" {
			continue
		}
		if _, duplicate := seen[identifier]; duplicate {
			return nil, fmt.Errorf("duplicate public DID verification method identifier %q", identifier)
		}
		seen[identifier] = struct{}{}
		identifiers = append(identifiers, identifier)
	}
	if len(identifiers) == 0 {
		return nil, errors.New("DID keystore has no public verification method identifiers")
	}
	return identifiers, nil
}

// DIDKeyMetadata preserves the single-key compatibility API used by the base
// upgrade scenario. Use DIDVerificationMethodIDs when a scenario creates more
// than one DID.
func (n *Network) DIDKeyMetadata(ctx context.Context, node *cosmos.ChainNode) (json.RawMessage, error) {
	identifiers, err := n.DIDVerificationMethodIDs(ctx, node)
	if err != nil {
		return nil, err
	}
	if len(identifiers) != 1 {
		return nil, fmt.Errorf("DID keystore has %d public verification method identifiers, want 1", len(identifiers))
	}
	metadata, err := json.Marshal(map[string]string{"address": identifiers[0]})
	if err != nil {
		return nil, fmt.Errorf("encode public DID key metadata: %w", err)
	}
	return metadata, nil
}
