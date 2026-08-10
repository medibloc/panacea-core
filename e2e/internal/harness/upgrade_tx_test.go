package harness

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLegacyDIDCreateCommandSuppliesMatchingNonEmptyPasswords(t *testing.T) {
	passwordFile := t.TempDir() + "/did-password"
	probe := []string{
		"sh", "-c",
		`IFS= read -r first
IFS= read -r second
[ "$first" = "$second" ]
[ "${#first}" -ge 8 ]`,
		"password-probe",
	}
	command := legacyDIDCreateCommand(passwordFile, probe)

	require.NotContains(t, strings.Join(command, " "), `printf '\\n\\n'`)
	require.NoError(t, exec.Command(command[0], command[1:]...).Run())
}

func TestLegacyDIDAuthenticatedCommandReusesNodeLocalPassword(t *testing.T) {
	passwordFile := t.TempDir() + "/did-password"
	createProbe := []string{
		"sh", "-c",
		`IFS= read -r first
IFS= read -r second
[ "$first" = "$second" ]`,
		"create-password-probe",
	}
	createCommand := legacyDIDCreateCommand(passwordFile, createProbe)
	require.NoError(t, exec.Command(createCommand[0], createCommand[1:]...).Run())

	authenticatedProbe := []string{
		"sh", "-c",
		`IFS= read -r supplied
IFS= read -r stored <"$1"
[ -n "$supplied" ]
[ "$supplied" = "$stored" ]`,
		"authenticated-password-probe",
		passwordFile,
	}
	authenticatedCommand := legacyDIDAuthenticatedCommand(passwordFile, authenticatedProbe)
	require.NoError(t, exec.Command(authenticatedCommand[0], authenticatedCommand[1:]...).Run())
}

func TestLegacyDIDCreateCommandRedactsMnemonicBeforeExecBoundary(t *testing.T) {
	const mnemonic = "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu"
	probe := []string{
		"sh", "-c",
		`printf '{"txhash":"ABC"}\n'
printf 'A random mnemonic was generated: %s\n' "$1" >&2
printf 'useful diagnostic\n' >&2
exit 17`,
		"redaction-probe", mnemonic,
	}
	arguments := legacyDIDCreateCommand(t.TempDir()+"/did-password", probe)
	command := exec.Command(arguments[0], arguments[1:]...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 17, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"txhash":"ABC"`)
	require.NotContains(t, stderr.String(), mnemonic)
	require.Contains(t, stderr.String(), legacyDIDMnemonicRedaction)
	require.Contains(t, stderr.String(), "useful diagnostic")
}

func TestLegacyDIDCreateCommandRedactsUnterminatedMnemonicLine(t *testing.T) {
	const mnemonic = "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu"
	probe := []string{
		"sh", "-c",
		`printf 'A random mnemonic was generated: %s' "$1" >&2`,
		"unterminated-redaction-probe", mnemonic,
	}
	arguments := legacyDIDCreateCommand(t.TempDir()+"/did-password", probe)
	command := exec.Command(arguments[0], arguments[1:]...)
	var stderr bytes.Buffer
	command.Stderr = &stderr

	require.NoError(t, command.Run())
	require.NotContains(t, stderr.String(), mnemonic)
	require.Equal(t, legacyDIDMnemonicRedaction, stderr.String())
}

func TestSanitizeLegacyDIDOutputRedactsEveryGeneratedMnemonic(t *testing.T) {
	const firstMnemonic = "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu"
	const secondMnemonic = "one two three four five six seven eight nine ten eleven twelve"
	output := []byte(
		"before\nA random mnemonic was generated: " + firstMnemonic +
			"\r\nmiddle\nA random mnemonic was generated: " + secondMnemonic + "\nafter\n",
	)

	sanitized := string(sanitizeLegacyDIDOutput(output))

	require.NotContains(t, sanitized, firstMnemonic)
	require.NotContains(t, sanitized, secondMnemonic)
	require.Equal(t, 2, strings.Count(sanitized, legacyDIDMnemonicRedaction))
	require.Contains(t, sanitized, "before")
	require.Contains(t, sanitized, "after")
}

func TestSanitizeLegacyDIDErrorRedactsMessageWithoutRetainingSecretCause(t *testing.T) {
	cause := errors.New("A random mnemonic was generated: secret mnemonic words")

	sanitized := sanitizeLegacyDIDError(cause)

	require.Error(t, sanitized)
	require.NotContains(t, sanitized.Error(), "secret mnemonic words")
	require.Contains(t, sanitized.Error(), legacyDIDMnemonicRedaction)
	require.False(t, errors.Is(sanitized, cause))
	require.NoError(t, sanitizeLegacyDIDError(nil))
}
