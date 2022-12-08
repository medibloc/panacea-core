package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

const (
	flagOracleUniqueID      = "oracle-unique-id"
	flagOraclePublicKey     = "oracle-public-key"
	flagOracleRemoteReport  = "oracle-remote-report"
	flagOraclePublicKeyPath = "oracle-public-key-path"

	flagOracleAccount  = "oracle-account"
	flagOracleEndpoint = "oracle-endpoint"
	flagOracleCommRate = "oracle-commission-rate"
)

type OraclePubKeyInfo struct {
	PublicKeyBase64    string `json:"public_key_base64"`
	RemoteReportBase64 string `json:"remote_report_base64"`
}

func AddGenesisOracleCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-oracle",
		Short: "Add a genesis oracle and oracle module parameters to genesis.json",
		Long: `
			Using this command, you can set genesis oracle and oracle module parameters.
			If you set oracle public key path for oracle module params, the flags for oracle public key and remote report would be ignored.

			The desired format of oracle public key file is:
			{
				"public_key_base64" : "<base64-encoded-oracle-public-key>",
				"remote_report_base64" : "<base64-encoded-oracle-remote-report>"
			}
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			oracleGenState := oracletypes.GetGenesisStateFromAppState(cdc, appState)

			if err := setOracle(cmd, oracleGenState); err != nil {
				return fmt.Errorf("failed to set oracle: %w", err)
			}

			if err := setOracleParams(cmd, oracleGenState); err != nil {
				return fmt.Errorf("failed to set oracle params: %w", err)
			}

			oracleGenStateBz, err := cdc.MarshalJSON(oracleGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal oracle genesis state: %w", err)
			}

			appState[oracletypes.ModuleName] = oracleGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flagOracleUniqueID, "", "unique ID for oracle")
	cmd.Flags().String(flagOraclePublicKey, "", "base64 encoded oracle public key")
	cmd.Flags().String(flagOracleRemoteReport, "", "base64 encoded remote report of oracle public key")
	cmd.Flags().String(flagOraclePublicKeyPath, "", "File path where 'oraclePublicKey' and 'oraclePublicKeyRemoteReport' are stored")
	cmd.Flags().String(flagOracleAccount, "", "address or keyName")
	cmd.Flags().String(flagOracleEndpoint, "", "oracle's endpoint")
	cmd.Flags().String(flagOracleCommRate, "", "oracle's commission rate")
	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func setOracle(cmd *cobra.Command, genState *oracletypes.GenesisState) error {
	clientCtx := client.GetClientContextFromCmd(cmd)

	oracleAddressOrKey, err := cmd.Flags().GetString(flagOracleAccount)
	if err != nil {
		return fmt.Errorf("failed to get oracle account flag: %w", err)
	}

	if len(oracleAddressOrKey) == 0 {
		return nil
	}

	oracleAccAddr, err := sdk.AccAddressFromBech32(oracleAddressOrKey)

	// if err is not nil, get address from key store
	if err != nil {
		inBuf := bufio.NewReader(cmd.InOrStdin())
		keyringBackend, err := cmd.Flags().GetString(flags.FlagKeyringBackend)
		if err != nil {
			return fmt.Errorf("failed to get keyring backend: %w", err)
		}

		kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf)
		if err != nil {
			return fmt.Errorf("failed to create new keyring instance: %w", err)
		}

		info, err := kb.Key(oracleAddressOrKey)
		if err != nil {
			return fmt.Errorf("failed to get address from Keybase: %w", err)
		}
		oracleAccAddr = info.GetAddress()
	}

	for _, oracle := range genState.Oracles {
		if oracle.OracleAddress == oracleAccAddr.String() {
			return fmt.Errorf("existing oracle. address: %s", oracleAccAddr.String())
		}
	}

	uniqueID, err := cmd.Flags().GetString(flagOracleUniqueID)
	if err != nil {
		return fmt.Errorf("failed to get oracle unique ID: %w", err)
	}

	endpoint, err := cmd.Flags().GetString(flagOracleEndpoint)
	if err != nil {
		return fmt.Errorf("failed to get oracle endpoint: %w", err)
	}

	commRateStr, err := cmd.Flags().GetString(flagOracleCommRate)
	if err != nil {
		return fmt.Errorf("failed to get oracle commission rate: %w", err)
	}

	commRate, err := sdk.NewDecFromStr(commRateStr)
	if err != nil {
		return fmt.Errorf("inavlid commission rate: %w", err)
	}

	if commRate.IsNegative() || commRate.GT(sdk.OneDec()) {
		return fmt.Errorf("oracle commission rate should be between 0 and 1")
	}

	genState.Oracles = append(genState.Oracles, oracletypes.Oracle{
		UniqueId:             uniqueID,
		OracleAddress:        oracleAccAddr.String(),
		Endpoint:             endpoint,
		OracleCommissionRate: commRate,
	})

	return nil
}

func setOracleParams(cmd *cobra.Command, genState *oracletypes.GenesisState) error {
	uniqueID, err := cmd.Flags().GetString(flagOracleUniqueID)
	if err != nil {
		return fmt.Errorf("falied to get unique ID: %w", err)
	}

	if len(uniqueID) > 0 {
		genState.Params.UniqueId = uniqueID
	}

	path, err := cmd.Flags().GetString(flagOraclePublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to get oracle public key path: %w", err)
	}

	if len(path) > 0 {
		contentBz, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read oracle public key path: %w", err)
		}

		var oraclePubKeyInfo OraclePubKeyInfo
		if err := json.Unmarshal(contentBz, &oraclePubKeyInfo); err != nil {
			return fmt.Errorf("failed to unmarshal oracle public key info: %w", err)
		}

		genState.Params.OraclePublicKey = oraclePubKeyInfo.PublicKeyBase64
		genState.Params.OraclePubKeyRemoteReport = oraclePubKeyInfo.RemoteReportBase64
	} else {
		pubKeyBase64, err := cmd.Flags().GetString(flagOraclePublicKey)
		if err != nil {
			return fmt.Errorf("failed to get oracle public key: %w", err)
		}

		genState.Params.OraclePublicKey = pubKeyBase64

		remoteReportBase64, err := cmd.Flags().GetString(flagOracleRemoteReport)
		if err != nil {
			return fmt.Errorf("failed to get oracle remote report: %w", err)
		}

		genState.Params.OraclePubKeyRemoteReport = remoteReportBase64
	}

	return genState.Params.Validate()
}
