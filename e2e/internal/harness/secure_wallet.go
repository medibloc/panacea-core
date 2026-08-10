package harness

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	ictcosmos "github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

var testWalletKeyNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

const walletSecretCleanupTimeout = 10 * time.Second

type secureWalletRecoveryPlan struct {
	RelativeSecretPath string
	Secret             []byte
	Command            []string
	CleanupCommand     []string
}

func newSecureWalletRecoveryPlan(
	homeDir string,
	binary string,
	coinType string,
	keyName string,
	mnemonic string,
	nonce int64,
) (secureWalletRecoveryPlan, error) {
	if !path.IsAbs(homeDir) {
		return secureWalletRecoveryPlan{}, errors.New("wallet recovery home must be absolute")
	}
	if strings.TrimSpace(binary) == "" {
		return secureWalletRecoveryPlan{}, errors.New("wallet recovery binary is required")
	}
	if _, err := strconv.ParseUint(coinType, 10, 32); err != nil {
		return secureWalletRecoveryPlan{}, fmt.Errorf("wallet recovery coin type: %w", err)
	}
	if !testWalletKeyNamePattern.MatchString(keyName) {
		return secureWalletRecoveryPlan{}, fmt.Errorf("unsafe test-wallet key name %q", keyName)
	}
	if strings.TrimSpace(mnemonic) == "" {
		return secureWalletRecoveryPlan{}, errors.New("wallet recovery mnemonic is required")
	}
	if nonce <= 0 {
		return secureWalletRecoveryPlan{}, errors.New("wallet recovery nonce must be positive")
	}

	relativeSecretPath := fmt.Sprintf(".e2e-secrets/recovery-%d.mnemonic", nonce)
	absoluteSecretPath := path.Join(homeDir, relativeSecretPath)
	return secureWalletRecoveryPlan{
		RelativeSecretPath: relativeSecretPath,
		Secret:             []byte(mnemonic + "\n"),
		Command: []string{
			"sh", "-c", `
set -eu
secret_path=$1
key_name=$2
binary=$3
coin_type=$4
home_dir=$5
trap 'rm -f "$secret_path"' 0 HUP INT TERM
"$binary" keys add "$key_name" --recover --keyring-backend test --coin-type "$coin_type" --home "$home_dir" --output json < "$secret_path"
`,
			"secure-wallet-recovery",
			absoluteSecretPath,
			keyName,
			binary,
			coinType,
			homeDir,
		},
		CleanupCommand: []string{"rm", "-f", absoluteSecretPath},
	}, nil
}

func executeSecureWalletRecovery(
	ctx context.Context,
	keyName string,
	plan secureWalletRecoveryPlan,
	env []string,
	writeFile func(context.Context, []byte, string) error,
	exec func(context.Context, []string, []string) ([]byte, []byte, error),
) (retErr error) {
	if writeFile == nil {
		return errors.New("wallet secret writer is required")
	}
	if exec == nil {
		return errors.New("wallet secret executor is required")
	}
	if len(plan.CleanupCommand) == 0 {
		return errors.New("wallet secret cleanup command is required")
	}

	// Register cleanup before staging. WriteFile can copy the file successfully
	// and still return an error, while the caller context can be canceled before
	// the recovery shell ever starts. A fresh bounded context gives the Docker
	// API an independent chance to remove the staged secret in both cases.
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), walletSecretCleanupTimeout)
		defer cancel()
		_, stderr, cleanupErr := exec(cleanupCtx, plan.CleanupCommand, env)
		if cleanupErr != nil {
			retErr = errors.Join(retErr, fmt.Errorf(
				"remove staged mnemonic for test wallet %q: %w: %s",
				keyName,
				cleanupErr,
				boundedString(stderr, txStderrMaxBytes),
			))
		}
	}()

	if err := writeFile(ctx, plan.Secret, plan.RelativeSecretPath); err != nil {
		return fmt.Errorf("stage mnemonic for test wallet %q: %w", keyName, err)
	}
	_, stderr, err := exec(ctx, plan.Command, env)
	if err != nil {
		return fmt.Errorf(
			"recover test wallet %q: %w: %s",
			keyName,
			err,
			boundedString(stderr, txStderrMaxBytes),
		)
	}
	return nil
}

// BuildWallet imports caller-supplied test mnemonics without embedding them in
// the command array that Interchaintest logs. The secret crosses the Docker API
// in a mode-0600 file, is consumed through stdin, and is removed by a shell
// trap. Empty mnemonics use Interchaintest's safe generated-key path.
func (n *Network) BuildWallet(ctx context.Context, keyName, mnemonic string) (ibc.Wallet, error) {
	if n == nil || n.Chain == nil {
		return nil, errors.New("test-wallet network is required")
	}
	if mnemonic == "" {
		return n.Chain.BuildWallet(ctx, keyName, "")
	}
	node := n.Chain.GetNode()
	if node == nil {
		return nil, errors.New("test-wallet node is required")
	}
	plan, err := newSecureWalletRecoveryPlan(
		node.HomeDir(),
		n.Chain.Config().Bin,
		n.Chain.Config().CoinType,
		keyName,
		mnemonic,
		time.Now().UTC().UnixNano(),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		for i := range plan.Secret {
			plan.Secret[i] = 0
		}
	}()
	if err := executeSecureWalletRecovery(
		ctx,
		keyName,
		plan,
		n.Chain.Config().Env,
		node.WriteFile,
		node.Exec,
	); err != nil {
		return nil, err
	}
	address, err := n.Chain.GetAddress(ctx, keyName)
	if err != nil {
		return nil, fmt.Errorf("read recovered test wallet %q address: %w", keyName, err)
	}
	return ictcosmos.NewWallet(keyName, address, mnemonic, n.Chain.Config()), nil
}
