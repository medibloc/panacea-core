package types

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestLegacyProtoGoPackageUsesModuleMajorVersion(t *testing.T) {
	descriptor, err := proto.HybridResolver.FindDescriptorByName(
		"panacea.pnft.v2.MsgCreateDenomRequest",
	)
	require.NoError(t, err)

	options, ok := descriptor.ParentFile().Options().(*descriptorpb.FileOptions)
	require.True(t, ok)
	require.Equal(
		t,
		"github.com/medibloc/panacea-core/v2/x/pnft/types",
		options.GetGoPackage(),
	)
}

func TestLegacyMsgAnyRoundTrip(t *testing.T) {
	registry := cdctypes.NewInterfaceRegistry()
	sdk.RegisterInterfaces(registry)
	RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	testCases := []struct {
		name      string
		msg       sdk.Msg
		typeURL   string
		binaryHex string
		json      string
	}{
		{
			name: "create denom",
			msg: &MsgCreateDenomRequest{
				Id:          "denom",
				Name:        "name",
				Symbol:      "symbol",
				Description: "description",
				Uri:         "https://example.com/denom.json",
				UriHash:     "denom-hash",
				Data:        "data",
				Creator:     "panacea1creator",
			},
			typeURL:   "/panacea.pnft.v2.MsgCreateDenomRequest",
			binaryHex: "0a262f70616e616365612e706e66742e76322e4d736743726561746544656e6f6d5265717565737412650a0564656e6f6d12046e616d651a0673796d626f6c220b6465736372697074696f6e2a1e68747470733a2f2f6578616d706c652e636f6d2f64656e6f6d2e6a736f6e320a64656e6f6d2d686173683a0464617461420f70616e616365613163726561746f72",
			json:      `{"@type":"/panacea.pnft.v2.MsgCreateDenomRequest","id":"denom","name":"name","symbol":"symbol","description":"description","uri":"https://example.com/denom.json","uri_hash":"denom-hash","data":"data","creator":"panacea1creator"}`,
		},
		{
			name: "update denom",
			msg: &MsgUpdateDenomRequest{
				Id:          "denom",
				Name:        "updated name",
				Symbol:      "updated symbol",
				Description: "updated description",
				Uri:         "https://example.com/updated-denom.json",
				UriHash:     "updated-denom-hash",
				Data:        "updated data",
				Updater:     "panacea1updater",
			},
			typeURL:   "/panacea.pnft.v2.MsgUpdateDenomRequest",
			binaryHex: "0a262f70616e616365612e706e66742e76322e4d736755706461746544656e6f6d526571756573741295010a0564656e6f6d120c75706461746564206e616d651a0e757064617465642073796d626f6c221375706461746564206465736372697074696f6e2a2668747470733a2f2f6578616d706c652e636f6d2f757064617465642d64656e6f6d2e6a736f6e3212757064617465642d64656e6f6d2d686173683a0c757064617465642064617461420f70616e616365613175706461746572",
			json:      `{"@type":"/panacea.pnft.v2.MsgUpdateDenomRequest","id":"denom","name":"updated name","symbol":"updated symbol","description":"updated description","uri":"https://example.com/updated-denom.json","uri_hash":"updated-denom-hash","data":"updated data","updater":"panacea1updater"}`,
		},
		{
			name: "delete denom",
			msg: &MsgDeleteDenomRequest{
				Id:      "denom",
				Remover: "panacea1remover",
			},
			typeURL:   "/panacea.pnft.v2.MsgDeleteDenomRequest",
			binaryHex: "0a262f70616e616365612e706e66742e76322e4d736744656c65746544656e6f6d5265717565737412180a0564656e6f6d120f70616e616365613172656d6f766572",
			json:      `{"@type":"/panacea.pnft.v2.MsgDeleteDenomRequest","id":"denom","remover":"panacea1remover"}`,
		},
		{
			name: "transfer denom",
			msg: &MsgTransferDenomRequest{
				Id:       "denom",
				Sender:   "panacea1sender",
				Receiver: "panacea1receiver",
			},
			typeURL:   "/panacea.pnft.v2.MsgTransferDenomRequest",
			binaryHex: "0a282f70616e616365612e706e66742e76322e4d73675472616e7366657244656e6f6d5265717565737412290a0564656e6f6d120e70616e616365613173656e6465721a1070616e61636561317265636569766572",
			json:      `{"@type":"/panacea.pnft.v2.MsgTransferDenomRequest","id":"denom","sender":"panacea1sender","receiver":"panacea1receiver"}`,
		},
		{
			name: "mint PNFT",
			msg: &MsgMintPNFTRequest{
				DenomId:     "denom",
				Id:          "pnft",
				Name:        "name",
				Description: "description",
				Uri:         "https://example.com/pnft.json",
				UriHash:     "pnft-hash",
				Data:        "data",
				Creator:     "panacea1creator",
			},
			typeURL:   "/panacea.pnft.v2.MsgMintPNFTRequest",
			binaryHex: "0a232f70616e616365612e706e66742e76322e4d73674d696e74504e46545265717565737412610a0564656e6f6d1204706e66741a046e616d65220b6465736372697074696f6e2a1d68747470733a2f2f6578616d706c652e636f6d2f706e66742e6a736f6e3209706e66742d686173683a0464617461420f70616e616365613163726561746f72",
			json:      `{"@type":"/panacea.pnft.v2.MsgMintPNFTRequest","denom_id":"denom","id":"pnft","name":"name","description":"description","uri":"https://example.com/pnft.json","uri_hash":"pnft-hash","data":"data","creator":"panacea1creator"}`,
		},
		{
			name: "transfer PNFT",
			msg: &MsgTransferPNFTRequest{
				DenomId:  "denom",
				Id:       "pnft",
				Sender:   "panacea1sender",
				Receiver: "panacea1receiver",
			},
			typeURL:   "/panacea.pnft.v2.MsgTransferPNFTRequest",
			binaryHex: "0a272f70616e616365612e706e66742e76322e4d73675472616e73666572504e465452657175657374122f0a0564656e6f6d1204706e66741a0e70616e616365613173656e646572221070616e61636561317265636569766572",
			json:      `{"@type":"/panacea.pnft.v2.MsgTransferPNFTRequest","denom_id":"denom","id":"pnft","sender":"panacea1sender","receiver":"panacea1receiver"}`,
		},
		{
			name: "burn PNFT",
			msg: &MsgBurnPNFTRequest{
				DenomId: "denom",
				Id:      "pnft",
				Burner:  "panacea1burner",
			},
			typeURL:   "/panacea.pnft.v2.MsgBurnPNFTRequest",
			binaryHex: "0a232f70616e616365612e706e66742e76322e4d73674275726e504e465452657175657374121d0a0564656e6f6d1204706e66741a0e70616e61636561316275726e6572",
			json:      `{"@type":"/panacea.pnft.v2.MsgBurnPNFTRequest","denom_id":"denom","id":"pnft","burner":"panacea1burner"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := cdctypes.NewAnyWithValue(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.typeURL, packed.TypeUrl)
			require.Equal(t, tc.typeURL, sdk.MsgTypeURL(tc.msg))

			t.Run("binary", func(t *testing.T) {
				binary, err := hex.DecodeString(tc.binaryHex)
				require.NoError(t, err)
				assertLegacyMsgAnyDecodes(t, registry, cdc, binary, tc.msg, tc.typeURL)

				encoded, err := cdc.Marshal(packed)
				require.NoError(t, err)
				require.Equal(t, binary, encoded)
			})

			t.Run("json", func(t *testing.T) {
				var decodedAny cdctypes.Any
				require.NoError(t, cdc.UnmarshalJSON([]byte(tc.json), &decodedAny))
				assertLegacyMsgUnpacked(t, registry, &decodedAny, tc.msg, tc.typeURL)

				encoded, err := cdc.MarshalJSON(packed)
				require.NoError(t, err)
				require.JSONEq(t, tc.json, string(encoded))
			})
		})
	}
}

func assertLegacyMsgAnyDecodes(
	t *testing.T,
	registry cdctypes.InterfaceRegistry,
	cdc *codec.ProtoCodec,
	binary []byte,
	expected sdk.Msg,
	expectedTypeURL string,
) {
	t.Helper()

	var decodedAny cdctypes.Any
	require.NoError(t, cdc.Unmarshal(binary, &decodedAny))
	assertLegacyMsgUnpacked(t, registry, &decodedAny, expected, expectedTypeURL)
}

func assertLegacyMsgUnpacked(
	t *testing.T,
	registry cdctypes.InterfaceRegistry,
	packed *cdctypes.Any,
	expected sdk.Msg,
	expectedTypeURL string,
) {
	t.Helper()

	require.Equal(t, expectedTypeURL, packed.TypeUrl)

	var decoded sdk.Msg
	require.NoError(t, registry.UnpackAny(packed, &decoded))
	require.IsType(t, expected, decoded)
	require.True(t, proto.Equal(expected, decoded))
}
