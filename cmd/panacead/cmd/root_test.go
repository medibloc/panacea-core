package cmd_test

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"testing"
	"time"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	panaceaapp "github.com/medibloc/panacea-core/v2/app"
	panaceacmd "github.com/medibloc/panacea-core/v2/cmd/panacead/cmd"
	"github.com/medibloc/panacea-core/v2/types/assets"
)

func TestNewRootCmdCommandTree(t *testing.T) {
	var root *cobra.Command
	require.NotPanics(t, func() {
		root, _ = panaceacmd.NewRootCmd()
	})
	require.NotNil(t, root)

	t.Run("query has one IBC transfer command", func(t *testing.T) {
		query := requireDirectChild(t, root, "query")
		requireDirectChild(t, query, "ibc-transfer")
	})

	t.Run("tx has one IBC transfer command", func(t *testing.T) {
		tx := requireDirectChild(t, root, "tx")
		requireDirectChild(t, tx, "ibc-transfer")
	})

	t.Run("keeps the legacy parameter change proposal command", func(t *testing.T) {
		requireCommandPath(t, root, "tx", "gov", "submit-legacy-proposal", "param-change")
	})

	t.Run("keeps the governance cancel proposal command", func(t *testing.T) {
		requireCommandPath(t, root, "tx", "gov", "cancel-proposal")
	})

	t.Run("keeps SDK genesis account command", func(t *testing.T) {
		command := requireCommandPath(t, root, "genesis", "add-genesis-account")
		for _, flagName := range []string{
			"append",
			"module-name",
			"vesting-amount",
			"vesting-end-time",
			"vesting-start-time",
		} {
			require.NotNil(t, command.Flags().Lookup(flagName), "missing flag %q", flagName)
		}
	})

	t.Run("keeps all Panacea module commands", func(t *testing.T) {
		testCases := []struct {
			path     []string
			expected []string
		}{
			{
				path:     []string{"query", "aol"},
				expected: []string{"get-topic", "list-topic", "get-writer", "list-writer", "get-record"},
			},
			{
				path:     []string{"tx", "aol"},
				expected: []string{"create-topic", "add-writer", "delete-writer", "add-record"},
			},
			{
				path:     []string{"query", "did"},
				expected: []string{"get-did"},
			},
			{
				path:     []string{"tx", "did"},
				expected: []string{"create-did", "update-did", "deactivate-did"},
			},
			{
				path: []string{"query", "nft"},
				expected: []string{
					"balance",
					"owner",
					"supply",
					"nfts",
					"nft",
					"class",
					"classes",
					"class-record",
					"nft-record",
					"nft-records",
				},
			},
			{
				path: []string{"tx", "nft"},
				expected: []string{
					"create-class",
					"update-controller",
					"mint",
					"revoke",
					"burn",
					"send",
				},
			},
		}

		for _, testCase := range testCases {
			moduleCommand := requireCommandPath(t, root, testCase.path...)
			actual := make([]string, 0, len(moduleCommand.Commands()))
			for _, command := range moduleCommand.Commands() {
				actual = append(actual, command.Name())
			}
			require.ElementsMatch(t, testCase.expected, actual, "unexpected command set at %v", testCase.path)
		}
	})

	t.Run("supports standard NFT owner-only list queries", func(t *testing.T) {
		nfts := requireCommandPath(t, root, "query", "nft", "nfts")
		require.NotNil(t, nfts.Flags().Lookup("owner"))
		require.NotNil(t, nfts.Args)
		require.NoError(t, nfts.Args(nfts, nil))
	})

	t.Run("keeps SDK module commands", func(t *testing.T) {
		requireCommandPath(t, root, "query", "auth", "account")
		requireCommandPath(t, root, "tx", "staking", "delegate")
	})

	t.Run("has unique command names", func(t *testing.T) {
		requireUniqueChildNames(t, root)
	})

	t.Run("init writes application genesis", func(t *testing.T) {
		home := t.TempDir()
		root.SetArgs([]string{
			"init",
			"test-node",
			"--home", home,
			"--chain-id", "test-chain",
			"--initial-height", "7",
		})
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)

		require.NoError(t, svrcmd.Execute(root, "", home))

		appConfigReader := viper.New()
		appConfigReader.SetConfigFile(filepath.Join(home, "config", "app.toml"))
		require.NoError(t, appConfigReader.ReadInConfig())
		appConfig, err := serverconfig.GetConfig(appConfigReader)
		require.NoError(t, err)
		require.Equal(t, uint64(10_000_000), appConfig.QueryGasLimit)
		require.Equal(t, uint(10), appConfig.API.RPCWriteTimeout)
		require.Equal(t, serverconfig.DefaultGRPCMaxRecvMsgSize, appConfig.GRPC.MaxRecvMsgSize)
		require.Equal(t, serverconfig.DefaultGRPCMaxRecvMsgSize, appConfig.GRPC.MaxSendMsgSize)

		appGenesis, err := genutiltypes.AppGenesisFromFile(filepath.Join(home, "config", "genesis.json"))
		require.NoError(t, err)
		require.Equal(t, "test-chain", appGenesis.ChainID)
		require.Equal(t, int64(7), appGenesis.InitialHeight)
		require.False(t, appGenesis.GenesisTime.IsZero())
		require.NotNil(t, appGenesis.Consensus)
		require.NotNil(t, appGenesis.Consensus.Params)
		require.Empty(t, appGenesis.Consensus.Validators)
		require.Equal(t, 21*24*time.Hour, appGenesis.Consensus.Params.Evidence.MaxAgeDuration)
		require.Equal(t, int64(21*24*time.Hour/(6*time.Second)), appGenesis.Consensus.Params.Evidence.MaxAgeNumBlocks)

		var appState map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(appGenesis.AppState, &appState))
		require.NotEmpty(t, appState)

		var govGenesis govv1types.GenesisState
		encodingConfig := panaceaapp.MakeEncodingConfig()
		require.NoError(t, encodingConfig.Codec.UnmarshalJSON(appState[govtypes.ModuleName], &govGenesis))
		require.NoError(t, govGenesis.Params.ValidateBasic())

		require.Len(t, govGenesis.Params.MinDeposit, 1)
		require.Len(t, govGenesis.Params.ExpeditedMinDeposit, 1)
		minDeposit := govGenesis.Params.MinDeposit[0]
		expeditedMinDeposit := govGenesis.Params.ExpeditedMinDeposit[0]
		require.Equal(t, assets.MicroMedDenom, minDeposit.Denom)
		require.Equal(t, minDeposit.Denom, expeditedMinDeposit.Denom)
		require.True(t, minDeposit.Amount.Equal(sdk.TokensFromConsensusPower(100000, sdk.DefaultPowerReduction)))
		require.True(t, expeditedMinDeposit.Amount.Equal(sdk.TokensFromConsensusPower(500000, sdk.DefaultPowerReduction)))
		require.True(t, expeditedMinDeposit.Amount.Equal(minDeposit.Amount.MulRaw(govv1types.DefaultMinExpeditedDepositTokensRatio)))

		require.NotNil(t, govGenesis.Params.VotingPeriod)
		require.Equal(t, 3*24*time.Hour, *govGenesis.Params.VotingPeriod)
		require.NotNil(t, govGenesis.Params.ExpeditedVotingPeriod)
		require.Equal(t, govv1types.DefaultExpeditedPeriod, *govGenesis.Params.ExpeditedVotingPeriod)
		require.Equal(t, govv1types.DefaultExpeditedThreshold.String(), govGenesis.Params.ExpeditedThreshold)

		require.Equal(t, "0.500000000000000000", govGenesis.Params.ProposalCancelRatio)
		require.Empty(t, govGenesis.Params.ProposalCancelDest)
	})
}

func TestNFTTransactionCommandsGenerateMessages(t *testing.T) {
	signer := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()
	receiver := sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String()

	testCases := []struct {
		name             string
		args             []string
		expectedTypeURL  string
		expectedSignerKV string
		expectedIDKV     string
	}{
		{
			name:             "Panacea revoke",
			args:             []string{"tx", "nft", "revoke", "class-id", "nft-id"},
			expectedTypeURL:  "/panacea.nft.v1.MsgRevokeRequest",
			expectedSignerKV: `"controller":"` + signer + `"`,
			expectedIDKV:     `"nft_id":"nft-id"`,
		},
		{
			name:             "standard send",
			args:             []string{"tx", "nft", "send", "class-id", "nft-id", receiver},
			expectedTypeURL:  "/cosmos.nft.v1beta1.MsgSend",
			expectedSignerKV: `"sender":"` + signer + `"`,
			expectedIDKV:     `"id":"nft-id"`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			root, _ := panaceacmd.NewRootCmd()
			output := new(bytes.Buffer)
			home := t.TempDir()

			root.SetArgs(append(testCase.args,
				"--from", signer,
				"--generate-only",
				"--chain-id", "test-chain",
				"--home", home,
			))
			root.SetOut(output)
			root.SetErr(io.Discard)

			require.NoError(t, svrcmd.Execute(root, "", home))
			require.Contains(t, output.String(), `"@type":"`+testCase.expectedTypeURL+`"`)
			require.Contains(t, output.String(), testCase.expectedSignerKV)
			require.Contains(t, output.String(), `"class_id":"class-id"`)
			require.Contains(t, output.String(), testCase.expectedIDKV)
		})
	}
}

func TestAutoCLIUnjailGeneratesTransaction(t *testing.T) {
	root, _ := panaceacmd.NewRootCmd()
	addressBytes := bytes.Repeat([]byte{1}, 20)
	account := sdk.AccAddress(addressBytes).String()
	validator := sdk.ValAddress(addressBytes).String()
	output := new(bytes.Buffer)
	home := t.TempDir()

	root.SetArgs([]string{
		"tx", "slashing", "unjail",
		"--from", account,
		"--generate-only",
		"--chain-id", "test-chain",
		"--home", home,
	})
	root.SetOut(output)
	root.SetErr(io.Discard)

	require.NoError(t, svrcmd.Execute(root, "", home))
	require.Contains(t, output.String(), `"@type":"/cosmos.slashing.v1beta1.MsgUnjail"`)
	require.Contains(t, output.String(), `"validator_addr":"`+validator+`"`)
}

func TestAutoCLIValidatorQueryRejectsInvalidAddress(t *testing.T) {
	root, _ := panaceacmd.NewRootCmd()
	home := t.TempDir()
	root.SetArgs([]string{
		"query", "staking", "validator", "not-a-validator-address",
		"--home", home,
	})
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	err := svrcmd.Execute(root, "", home)
	require.ErrorContains(t, err, "decoding bech32 failed")
}

func requireUniqueChildNames(t *testing.T, parent *cobra.Command) {
	t.Helper()

	seen := make(map[string]struct{}, len(parent.Commands()))
	for _, child := range parent.Commands() {
		_, duplicate := seen[child.Name()]
		require.False(t, duplicate, "duplicate command %q under %q", child.Name(), parent.CommandPath())

		seen[child.Name()] = struct{}{}
		requireUniqueChildNames(t, child)
	}
}

func requireCommandPath(t *testing.T, root *cobra.Command, path ...string) *cobra.Command {
	t.Helper()

	current := root
	for _, name := range path {
		current = requireDirectChild(t, current, name)
	}
	return current
}

func requireDirectChild(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
	t.Helper()

	var matches []*cobra.Command
	for _, child := range parent.Commands() {
		if child.Name() == name {
			matches = append(matches, child)
		}
	}

	require.Len(t, matches, 1, "expected exactly one %q command under %q", name, parent.CommandPath())
	return matches[0]
}
