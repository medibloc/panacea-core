package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

const (
	flagUniqueID           = "oracle-unique-id"
	flagOraclePublicKey    = "oracle-public-key"
	flagOracleRemoteReport = "oracle-remote-report"
	flagOracleAccount      = "oracle-account"
)

func AddGenesisOracleCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-oracle",
		Short: "Add a genesis oracle to genesis.json",
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
				return err
			}

			if err := setOracleParams(cmd, oracleGenState); err != nil {
				return err
			}

			oracleGenStateBz, err := cdc.MarshalJSON(oracleGenState)
			if err != nil {
				return err
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
	cmd.Flags().String(flagUniqueID, "", "oracle's uniqueID")
	cmd.Flags().String(flagOraclePublicKey, "", "base64 encoded oracle public key")
	cmd.Flags().String(flagOracleRemoteReport, "", "base64 encoded remoteReport with oracle public key")
	cmd.Flags().String(flagOracleAccount, "", "address or keyName")
	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func setOracle(cmd *cobra.Command, genState *oracletypes.GenesisState) error {
	clientCtx := client.GetClientContextFromCmd(cmd)

	oracleAddressOrKey, err := cmd.Flags().GetString(flagOracleAccount)
	if err != nil {
		return err
	}

	if len(oracleAddressOrKey) > 0 {
		oracleAccAddr, err := sdk.AccAddressFromBech32(oracleAddressOrKey)
		if err != nil {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			keyringBackend, err := cmd.Flags().GetString(flags.FlagKeyringBackend)
			if err != nil {
				return err
			}

			// attempt to lookup address from Keybase if no address was provided
			kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf)
			if err != nil {
				return err
			}

			info, err := kb.Key(oracleAddressOrKey)
			if err != nil {
				return fmt.Errorf("failed to get address from Keybase: %w", err)
			}
			oracleAccAddr = info.GetAddress()
		}

		for _, oracle := range genState.Oracles {
			if oracle.Address == oracleAccAddr.String() {
				return fmt.Errorf("already exist oracle. address: %s", oracle.Address)
			}
		}

		genState.Oracles = append(genState.Oracles, oracletypes.Oracle{
			Address: oracleAccAddr.String(),
			Status:  oracletypes.ORACLE_STATUS_ACTIVE,
		})
	}

	return nil
}

// setOracleParams sets oraclePublicKey, oraclePubKeyRemoteReport, uniqueID existing in params of oracle module
func setOracleParams(cmd *cobra.Command, genState *oracletypes.GenesisState) error {
	uniqueID, err := cmd.Flags().GetString(flagUniqueID)
	if err != nil {
		return err
	}
	if len(uniqueID) > 0 {
		genState.Params.UniqueId = uniqueID
	}

	oraclePublicKeyBz, err := getBytesFromBase64(cmd, flagOraclePublicKey)
	if err != nil {
		return err
	}

	if len(oraclePublicKeyBz) > 0 {
		if _, err = btcec.ParsePubKey(oraclePublicKeyBz, btcec.S256()); err != nil {
			return err
		}
		genState.Params.OraclePublicKey = oraclePublicKeyBz
	}

	reportBz, err := getBytesFromBase64(cmd, flagOracleRemoteReport)
	if err != nil {
		return err
	}

	if len(reportBz) > 0 {
		genState.Params.OraclePubKeyRemoteReport = reportBz
	}

	return nil
}

func getBytesFromBase64(cmd *cobra.Command, key string) ([]byte, error) {
	content, err := cmd.Flags().GetString(key)
	if err != nil {
		return nil, err
	}

	if len(content) > 0 {
		contentBz, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, err
		}

		return contentBz, nil
	}

	return nil, nil
}