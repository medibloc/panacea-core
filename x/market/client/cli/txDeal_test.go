package cli
//
//import (
//	"fmt"
//	"github.com/cosmos/cosmos-sdk/client"
//	"github.com/cosmos/cosmos-sdk/client/flags"
//	"github.com/cosmos/cosmos-sdk/crypto/hd"
//	"github.com/cosmos/cosmos-sdk/crypto/keyring"
//	"github.com/cosmos/cosmos-sdk/testutil"
//	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
//	"github.com/cosmos/cosmos-sdk/testutil/network"
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
//	"github.com/gogo/protobuf/proto"
//	"github.com/medibloc/panacea-core/v2/app"
//	"github.com/medibloc/panacea-core/v2/types/testsuite"
//	cli "github.com/medibloc/panacea-core/v2/x/market/client/cli"
//	"github.com/medibloc/panacea-core/v2/x/market/types"
//	"github.com/stretchr/testify/suite"
//	"testing"
//)
//
//type cmdTestSuite struct {
//	testsuite.TestSuite
//
//	cfg     network.Config
//	network *network.Network
//}
//
//func (suite *cmdTestSuite) SetupSuite() {
//	suite.T().Log("setting up tx test suite")
//
//	suite.cfg = app.DefaultConfig()
//
//	genesisState := app.ModuleBasics.DefaultGenesis(suite.cfg.Codec)
//	marketGen := types.DefaultGenesis()
//	marketGenJSON := suite.cfg.Codec.MustMarshalJSON(marketGen)
//	genesisState[types.ModuleName] = marketGenJSON
//	suite.cfg.GenesisState = genesisState
//
//	suite.network = network.New(suite.T(), suite.cfg)
//
//	_, err := suite.network.WaitForHeight(1)
//	suite.Require().NoError(err)
//
//	val := suite.network.Validators[0]
//
//	_, err = MsgCreateDeal(suite.T(), val.ClientCtx, val.Address)
//	suite.Require().NoError(err)
//
//	_, err = suite.network.WaitForHeight(1)
//	suite.Require().NoError(err)
//}
//
//func (suite *cmdTestSuite) TearDownSuite() {
//	suite.T().Log("tearing down cmd test suite")
//	suite.network.Cleanup()
//}
//
//func (suite *cmdTestSuite) TestCmdCreateDeal() {
//	val := suite.network.Validators[0]
//
//	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewCreateDealAddr",
//		keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
//	suite.Require().NoError(err)
//
//	newCreatorAddr := sdk.AccAddress(info.GetPubKey().Address())
//
//	_, err = banktestutil.MsgSendExec(
//		val.ClientCtx,
//		val.Address,
//		newCreatorAddr,
//		sdk.NewCoins(sdk.NewInt64Coin(suite.cfg.BondDenom, 2000000), sdk.NewInt64Coin("node0token", 20000)),
//		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//	)
//	suite.Require().NoError(err)
//
//	testCases := []struct {
//		name         string
//		json         string
//		expectErr    bool
//		resType      proto.Message
//		expectedCode uint32
//	}{
//		{
//			"create deal cmd",
//			fmt.Sprintf(`
//			{
//			  "data_schema": [
//				"https://xxx.jsonld"
//			  ],
//			  "budget": "2stake,100node0token",
//			  "max_num_data": 10000,
//			  "trusted_data_validators": [
//				"%s"
//			  ]
//			}
//			`, val.Address), false, &sdk.TxResponse{}, 0,
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//
//		suite.Run(tc.name, func() {
//			cmd := cli.CmdCreateDeal()
//			clientCtx := val.ClientCtx
//
//			jsonFile := testutil.WriteToNewTempFile(suite.T(), tc.json)
//
//			args := []string{
//				fmt.Sprintf("--%s=%s", cli.FlagDealFile, jsonFile.Name()),
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
//				// common args
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagGas, fmt.Sprint(300000)),
//			}
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
//			if tc.expectErr {
//				suite.Require().Error(err)
//				fmt.Println(err)
//			} else {
//				suite.Require().NoError(err)
//				err = clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.resType)
//				suite.Require().NoError(err)
//
//				txResponse := tc.resType.(*sdk.TxResponse)
//				suite.Require().Equal(tc.expectedCode, txResponse.Code, out.String())
//			}
//		})
//	}
//}
//
////func (suite *cmdTestSuite) TestCmdSellData() {
////	val := suite.network.Validators[0]
////
////	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewSellDataAddr",
////		keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
////	suite.Require().NoError(err)
////
////	newSellerAddress := sdk.AccAddress(info.GetPubKey().Address())
////
////	_, err = banktestutil.MsgSendExec(
////		val.ClientCtx,
////		val.Address,
////		newSellerAddress,
////		sdk.NewCoins(sdk.NewInt64Coin(suite.cfg.BondDenom, 200000000), sdk.NewInt64Coin(assets.MicroMedDenom, 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
////		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
////	)
////	suite.Require().NoError(err)
////
////	testCase := []struct {
////		name         string
////		json         string
////		expectErr    bool
////		resType      proto.Message
////		expectedCode uint32
////	}{
////		{
////			"sell data",
////			fmt.Sprintf(`
////				{
////				  "certificate": {
////					"unsigned_cert": {
////					  "deal_id": 1,
////					  "data_hash": "12e3d12fc312x3",
////					  "encrypted_data_url": "https://panacea.org/123.json",
////					  "data_validator_address": "%s",
////					  "requester_address": "%s"
////					},
////					"signature": "0x214abkdj3lfdsf3f0"
////				  }
////				}
////		`, val.Address, newSellerAddress), true, &sdk.TxResponse{}, 4,
////		},
////	}
////
////	for _, tc := range testCase {
////		tc := tc
////
////		suite.Run(tc.name, func() {
////			cmd := cli.CmdSellData()
////			clientCtx := val.ClientCtx
////
////			jsonFile := testutil.WriteToNewTempFile(suite.T(), tc.json)
////
////			args := []string{
////				fmt.Sprintf("--%s=%s", cli.DataVerificationCertificateFile, jsonFile.Name()),
////				fmt.Sprintf("--%s=%s", flags.FlagFrom, newSellerAddress),
////				// common args
////				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
////				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
////				fmt.Sprintf("--%s=%s", flags.FlagGas, fmt.Sprint(300000)),
////			}
////
////			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
////			if tc.expectErr {
////				suite.Require().Error(err)
////				fmt.Println(err)
////			} else {
////				suite.Require().NoError(err)
////				err = clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.resType)
////				suite.Require().NoError(err)
////
////				txResponse := tc.resType.(*sdk.TxResponse)
////				suite.Require().Equal(tc.expectedCode, txResponse.Code, out.String())
////			}
////		})
////
////	}
////}
//
//var commonArgs = []string{
//	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
//}
//
//func MsgCreateDeal(t *testing.T, clientCtx client.Context, owner fmt.Stringer) (testutil.BufferWriter, error) {
//	args := []string{}
//
//	jsonFile := testutil.WriteToNewTempFile(t,
//		fmt.Sprintf(
//			`
//		{
//		  "data_schema": [
//			"https://xxx.jsonld"
//		  ],
//		  "budget": "1000000node0token",
//		  "max_num_data": 10000,
//		  "trusted_data_validators": [
//			"%s"
//		  ]
//		}
//		`, owner.String()),
//	)
//
//	args = append(args,
//		fmt.Sprintf("--%s=%s", cli.FlagDealFile, jsonFile.Name()),
//		fmt.Sprintf("--%s=%s", flags.FlagFrom, owner.String()),
//		fmt.Sprintf("--%s=%d", flags.FlagGas, 300000),
//	)
//
//	args = append(args, commonArgs...)
//	return clitestutil.ExecTestCLICmd(clientCtx, cli.CmdCreateDeal(), args)
//}
//
//func TestDealTestSuite(t *testing.T) {
//	suite.Run(t, new(cmdTestSuite))
//}
