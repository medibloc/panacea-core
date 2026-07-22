package types

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	errorsmod "cosmossdk.io/errors"
	upstreamnft "cosmossdk.io/x/nft"
	cosmosproto "github.com/cosmos/cosmos-proto"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/gogoproto/proto"
	"github.com/medibloc/panacea-core/v2/app/params"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestModuleConstants(t *testing.T) {
	require.Equal(t, upstreamnft.ModuleName, ModuleName)
	require.Equal(t, upstreamnft.StoreKey, StoreKey)
	require.Equal(t, "policy_nft", PolicyStoreKey)
	require.Equal(t, "nftpolicy", PolicyCodespace)
	require.Equal(t, ModuleName, RouterKey)
	require.Equal(t, ModuleName, QuerierRoute)
	require.Equal(t, "/panacea.nft.v1.BasicNFTData", BasicNFTDataTypeURL)
	require.Equal(t, "panacea.nft.v1.NFTData", NFTDataInterfaceName)
}

func TestSentinelErrorContract(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		code uint32
	}{
		{name: "transfer not allowed", err: ErrTransferNotAllowed, code: 2},
		{name: "revocation not allowed", err: ErrRevocationNotAllowed, code: 3},
		{name: "nft revoked", err: ErrNFTRevoked, code: 4},
		{name: "max supply reached", err: ErrMaxSupplyReached, code: 5},
		{name: "nft id permanently used", err: ErrNFTIDPermanentlyUsed, code: 6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			codespace, code, _ := errorsmod.ABCIInfo(tc.err, false)
			require.Equal(t, PolicyCodespace, codespace)
			require.Equal(t, tc.code, code)
		})
	}
}

func TestRegisterInterfacesMessageRoundTrip(t *testing.T) {
	registry, cdc := newTestCodec()

	testCases := []struct {
		name    string
		msg     sdk.Msg
		typeURL string
	}{
		{
			name: "create class",
			msg: &MsgCreateClassRequest{
				Creator:        "panacea1creator",
				LocalClassId:   "certificate",
				Name:           "Certificate",
				Symbol:         "CERT",
				Description:    "A certificate class",
				Uri:            "https://example.com/class.json",
				UriHash:        "class-hash",
				TransferPolicy: TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
				Revocable:      true,
				MaxSupply:      100,
			},
			typeURL: "/panacea.nft.v1.MsgCreateClassRequest",
		},
		{
			name: "update controller",
			msg: &MsgUpdateControllerRequest{
				ClassId:       "panacea1creator:certificate",
				Controller:    "panacea1controller",
				NewController: "panacea1newcontroller",
			},
			typeURL: "/panacea.nft.v1.MsgUpdateControllerRequest",
		},
		{
			name: "mint",
			msg: &MsgMintRequest{
				ClassId:    "panacea1creator:certificate",
				NftId:      "nft-1",
				Controller: "panacea1controller",
				Recipient:  "panacea1recipient",
				Uri:        "https://example.com/nft-1.json",
				UriHash:    "nft-hash",
			},
			typeURL: "/panacea.nft.v1.MsgMintRequest",
		},
		{
			name: "revoke",
			msg: &MsgRevokeRequest{
				ClassId:    "panacea1creator:certificate",
				NftId:      "nft-1",
				Controller: "panacea1controller",
			},
			typeURL: "/panacea.nft.v1.MsgRevokeRequest",
		},
		{
			name: "burn",
			msg: &MsgBurnRequest{
				ClassId: "panacea1creator:certificate",
				NftId:   "nft-1",
				Owner:   "panacea1owner",
			},
			typeURL: "/panacea.nft.v1.MsgBurnRequest",
		},
		{
			name: "standard send",
			msg: &upstreamnft.MsgSend{
				ClassId:  "panacea1creator:certificate",
				Id:       "nft-1",
				Sender:   "panacea1sender",
				Receiver: "panacea1receiver",
			},
			typeURL: "/cosmos.nft.v1beta1.MsgSend",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := cdctypes.NewAnyWithValue(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.typeURL, packed.TypeUrl)

			binary, err := cdc.Marshal(packed)
			require.NoError(t, err)

			var decodedAny cdctypes.Any
			require.NoError(t, cdc.Unmarshal(binary, &decodedAny))

			var decoded sdk.Msg
			require.NoError(t, registry.UnpackAny(&decodedAny, &decoded))
			require.IsType(t, tc.msg, decoded)
			require.True(t, proto.Equal(tc.msg, decoded))
		})
	}
}

func TestLegacyAminoJSONSignModeWithNFTData(t *testing.T) {
	encodingConfig := params.MakeEncodingConfig()
	RegisterInterfaces(encodingConfig.InterfaceRegistry)

	data, err := cdctypes.NewAnyWithValue(&BasicNFTData{
		Name:        "Certificate #1",
		Description: "Completion certificate",
		ImageUri:    "https://example.com/nft-1.png",
	})
	require.NoError(t, err)

	msg := &MsgMintRequest{
		ClassId:    "panacea1creator:certificate",
		NftId:      "nft-1",
		Controller: "panacea1controller",
		Recipient:  "panacea1recipient",
		Uri:        "https://example.com/nft-1.json",
		UriHash:    "nft-hash",
		Data:       data,
	}
	binary, err := encodingConfig.Codec.Marshal(msg)
	require.NoError(t, err)
	var decodedMsg MsgMintRequest
	require.NoError(t, encodingConfig.Codec.Unmarshal(binary, &decodedMsg))
	require.IsType(t, &BasicNFTData{}, decodedMsg.Data.GetCachedValue())

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("umed", 10)))
	txBuilder.SetGasLimit(200000)
	txBuilder.SetMemo("memo")
	txBuilder.SetTimeoutHeight(123)
	require.NoError(t, txBuilder.SetMsgs(&decodedMsg))

	signBytes, err := authsigning.GetSignBytesAdapter(
		context.Background(),
		encodingConfig.TxConfig.SignModeHandler(),
		signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		authsigning.SignerData{
			Address:       "panacea1controller",
			ChainID:       "test-chain",
			AccountNumber: 7,
			Sequence:      11,
		},
		txBuilder.GetTx(),
	)
	require.NoError(t, err)

	var expectedSignBytes bytes.Buffer
	require.NoError(t, json.Compact(&expectedSignBytes, []byte(`
		{
			"account_number": "7",
			"chain_id": "test-chain",
			"fee": {
				"amount": [{"amount": "10", "denom": "umed"}],
				"gas": "200000"
			},
			"memo": "memo",
			"msgs": [{
				"type": "/panacea.nft.v1.MsgMintRequest",
				"value": {
					"class_id": "panacea1creator:certificate",
					"controller": "panacea1controller",
					"data": {
						"type": "/panacea.nft.v1.BasicNFTData",
						"value": {
							"description": "Completion certificate",
							"image_uri": "https://example.com/nft-1.png",
							"name": "Certificate #1"
						}
					},
					"nft_id": "nft-1",
					"recipient": "panacea1recipient",
					"uri": "https://example.com/nft-1.json",
					"uri_hash": "nft-hash"
				}
			}],
			"sequence": "11",
			"timeout_height": "123"
		}
	`)))
	require.Equal(t, expectedSignBytes.Bytes(), signBytes)
}

func TestRegisterInterfacesMsgServiceResponses(t *testing.T) {
	registry, _ := newTestCodec()

	testCases := []struct {
		typeURL string
		want    proto.Message
	}{
		{typeURL: "/panacea.nft.v1.MsgCreateClassResponse", want: &MsgCreateClassResponse{}},
		{typeURL: "/panacea.nft.v1.MsgUpdateControllerResponse", want: &MsgUpdateControllerResponse{}},
		{typeURL: "/panacea.nft.v1.MsgMintResponse", want: &MsgMintResponse{}},
		{typeURL: "/panacea.nft.v1.MsgRevokeResponse", want: &MsgRevokeResponse{}},
		{typeURL: "/panacea.nft.v1.MsgBurnResponse", want: &MsgBurnResponse{}},
	}

	for _, tc := range testCases {
		resolved, err := registry.Resolve(tc.typeURL)
		require.NoError(t, err)
		require.IsType(t, tc.want, resolved)
	}
}

func TestRegisterInterfacesNFTData(t *testing.T) {
	registry, _ := newTestCodec()

	require.Contains(t, registry.ListAllInterfaces(), NFTDataInterfaceName)
	require.Equal(
		t,
		[]string{BasicNFTDataTypeURL},
		registry.ListImplementations(NFTDataInterfaceName),
	)
}

func TestNFTDataProtoAnnotations(t *testing.T) {
	interfacesFile, err := proto.HybridResolver.FindFileByPath("panacea/nft/v1/interfaces.proto")
	require.NoError(t, err)
	declarations := protov2.GetExtension(
		interfacesFile.Options(),
		cosmosproto.E_DeclareInterface,
	).([]*cosmosproto.InterfaceDescriptor)
	require.Len(t, declarations, 1)
	require.Equal(t, "NFTData", declarations[0].GetName())
	require.Equal(t, "Metadata accepted by the Panacea NFT module.", declarations[0].GetDescription())

	basicData := messageDescriptor(t, "panacea.nft.v1.BasicNFTData")
	implementations := protov2.GetExtension(
		basicData.Options(),
		cosmosproto.E_ImplementsInterface,
	).([]string)
	require.Equal(t, []string{NFTDataInterfaceName}, implementations)

	assertAcceptsNFTData(t, messageDescriptor(t, "panacea.nft.v1.BurnTombstone"))
	assertAcceptsNFTData(t, messageDescriptor(t, "panacea.nft.v1.MsgMintRequest"))
}

func TestBasicNFTDataAnyRoundTrip(t *testing.T) {
	registry, cdc := newTestCodec()
	original := &BasicNFTData{
		Name:        "Certificate #1",
		Description: "Completion certificate",
		ImageUri:    "https://example.com/nft-1.png",
	}

	packed, err := cdctypes.NewAnyWithValue(original)
	require.NoError(t, err)
	require.Equal(t, BasicNFTDataTypeURL, packed.TypeUrl)

	t.Run("binary", func(t *testing.T) {
		binary, err := cdc.Marshal(packed)
		require.NoError(t, err)

		var decodedAny cdctypes.Any
		require.NoError(t, cdc.Unmarshal(binary, &decodedAny))
		require.Equal(t, BasicNFTDataTypeURL, decodedAny.TypeUrl)

		var decoded NFTData
		require.NoError(t, registry.UnpackAny(&decodedAny, &decoded))
		require.Equal(t, original, decoded)
	})

	t.Run("json", func(t *testing.T) {
		jsonBytes, err := cdc.MarshalJSON(packed)
		require.NoError(t, err)

		var decodedAny cdctypes.Any
		require.NoError(t, cdc.UnmarshalJSON(jsonBytes, &decodedAny))
		require.Equal(t, BasicNFTDataTypeURL, decodedAny.TypeUrl)

		var decoded NFTData
		require.NoError(t, registry.UnpackAny(&decodedAny, &decoded))
		require.Equal(t, original, decoded)
	})
}

func TestMsgMintRequestUnpacksNFTData(t *testing.T) {
	_, cdc := newTestCodec()
	data, err := cdctypes.NewAnyWithValue(&BasicNFTData{Name: "metadata"})
	require.NoError(t, err)
	request := &MsgMintRequest{Data: data}

	binary, err := cdc.Marshal(request)
	require.NoError(t, err)
	var decoded MsgMintRequest
	require.NoError(t, cdc.Unmarshal(binary, &decoded))
	require.IsType(t, &BasicNFTData{}, decoded.Data.GetCachedValue())

	unknown := &MsgMintRequest{Data: &cdctypes.Any{
		TypeUrl: "/panacea.nft.v1.UnknownNFTData",
		Value:   data.Value,
	}}
	binary, err = cdc.Marshal(unknown)
	require.NoError(t, err)
	err = cdc.Unmarshal(binary, &decoded)
	require.ErrorContains(t, err, "no concrete type registered for type URL")
}

func TestBurnTombstoneContainersUnpackNFTData(t *testing.T) {
	_, cdc := newTestCodec()
	data, err := cdctypes.NewAnyWithValue(&BasicNFTData{Name: "metadata"})
	require.NoError(t, err)
	tombstone := &BurnTombstone{Data: data}

	binary, err := cdc.Marshal(tombstone)
	require.NoError(t, err)
	var decodedTombstone BurnTombstone
	require.NoError(t, cdc.Unmarshal(binary, &decodedTombstone))
	require.IsType(t, &BasicNFTData{}, decodedTombstone.Data.GetCachedValue())

	genesis := &GenesisState{
		NftState:   upstreamnft.DefaultGenesisState(),
		Tombstones: []*BurnTombstone{tombstone},
	}
	binary, err = cdc.Marshal(genesis)
	require.NoError(t, err)
	var decodedGenesis GenesisState
	require.NoError(t, cdc.Unmarshal(binary, &decodedGenesis))
	require.IsType(
		t,
		&BasicNFTData{},
		decodedGenesis.Tombstones[0].Data.GetCachedValue(),
	)

	response := &QueryNFTRecordResponse{
		NftRecord: &NFTRecord{
			Record: &NFTRecord_BurnTombstone{BurnTombstone: tombstone},
		},
	}
	binary, err = cdc.Marshal(response)
	require.NoError(t, err)
	var decodedResponse QueryNFTRecordResponse
	require.NoError(t, cdc.Unmarshal(binary, &decodedResponse))
	require.IsType(
		t,
		&BasicNFTData{},
		decodedResponse.NftRecord.GetBurnTombstone().Data.GetCachedValue(),
	)

	unknown := &BurnTombstone{Data: &cdctypes.Any{
		TypeUrl: "/panacea.nft.v1.UnknownNFTData",
		Value:   data.Value,
	}}
	binary, err = cdc.Marshal(unknown)
	require.NoError(t, err)
	err = cdc.Unmarshal(binary, &decodedTombstone)
	require.ErrorContains(t, err, "no concrete type registered for type URL")
}

func TestLiveNFTContainersUnpackNFTData(t *testing.T) {
	_, cdc := newTestCodec()
	data, err := cdctypes.NewAnyWithValue(&BasicNFTData{Name: "metadata"})
	require.NoError(t, err)
	live := &LiveNFTRecord{
		Nft: &upstreamnft.NFT{Data: data},
	}

	binary, err := cdc.Marshal(live)
	require.NoError(t, err)
	var decodedLive LiveNFTRecord
	require.NoError(t, cdc.Unmarshal(binary, &decodedLive))
	require.IsType(t, &BasicNFTData{}, decodedLive.Nft.Data.GetCachedValue())

	pointResponse := &QueryNFTRecordResponse{
		NftRecord: &NFTRecord{
			Record: &NFTRecord_Live{Live: live},
		},
	}
	binary, err = cdc.Marshal(pointResponse)
	require.NoError(t, err)
	var decodedPointResponse QueryNFTRecordResponse
	require.NoError(t, cdc.Unmarshal(binary, &decodedPointResponse))
	require.IsType(
		t,
		&BasicNFTData{},
		decodedPointResponse.NftRecord.GetLive().Nft.Data.GetCachedValue(),
	)

	listResponse := &QueryNFTRecordsResponse{
		NftRecords: []*LiveNFTRecord{live},
	}
	binary, err = cdc.Marshal(listResponse)
	require.NoError(t, err)
	var decodedListResponse QueryNFTRecordsResponse
	require.NoError(t, cdc.Unmarshal(binary, &decodedListResponse))
	require.IsType(
		t,
		&BasicNFTData{},
		decodedListResponse.NftRecords[0].Nft.Data.GetCachedValue(),
	)

	genesis := &GenesisState{
		NftState: &upstreamnft.GenesisState{
			Entries: []*upstreamnft.Entry{
				{Nfts: []*upstreamnft.NFT{{Data: data}}},
			},
		},
	}
	binary, err = cdc.Marshal(genesis)
	require.NoError(t, err)
	var decodedGenesis GenesisState
	require.NoError(t, cdc.Unmarshal(binary, &decodedGenesis))
	require.IsType(
		t,
		&BasicNFTData{},
		decodedGenesis.NftState.Entries[0].Nfts[0].Data.GetCachedValue(),
	)
}

func TestUnknownNFTDataTypeURLFailsToDecode(t *testing.T) {
	registry, cdc := newTestCodec()
	unknown := &cdctypes.Any{
		TypeUrl: "/panacea.nft.v1.UnknownNFTData",
		Value:   []byte{0x0a, 0x00},
	}

	var decoded NFTData
	err := registry.UnpackAny(unknown, &decoded)
	require.ErrorContains(t, err, "no concrete type registered for type URL")

	var decodedJSON cdctypes.Any
	err = cdc.UnmarshalJSON(
		[]byte(`{"@type":"/panacea.nft.v1.UnknownNFTData","name":"unknown"}`),
		&decodedJSON,
	)
	require.ErrorContains(t, err, "unable to resolve type URL")
}

func TestBasicNFTDataIsNotRegisteredAsMessage(t *testing.T) {
	registry, cdc := newTestCodec()
	packed, err := cdctypes.NewAnyWithValue(&BasicNFTData{Name: "metadata"})
	require.NoError(t, err)

	binary, err := cdc.Marshal(packed)
	require.NoError(t, err)
	var decodedAny cdctypes.Any
	require.NoError(t, cdc.Unmarshal(binary, &decodedAny))

	var decoded sdk.Msg
	err = registry.UnpackAny(&decodedAny, &decoded)
	require.ErrorContains(t, err, "no concrete type registered for type URL")
}

func newTestCodec() (cdctypes.InterfaceRegistry, *codec.ProtoCodec) {
	registry := cdctypes.NewInterfaceRegistry()
	sdk.RegisterInterfaces(registry)
	RegisterInterfaces(registry)

	return registry, codec.NewProtoCodec(registry)
}

func messageDescriptor(t *testing.T, name protoreflect.FullName) protoreflect.MessageDescriptor {
	t.Helper()

	descriptor, err := proto.HybridResolver.FindDescriptorByName(name)
	require.NoError(t, err)
	message, ok := descriptor.(protoreflect.MessageDescriptor)
	require.True(t, ok)

	return message
}

func assertAcceptsNFTData(t *testing.T, message protoreflect.MessageDescriptor) {
	t.Helper()

	dataField := message.Fields().ByName("data")
	require.NotNil(t, dataField)
	acceptedInterface := protov2.GetExtension(
		dataField.Options(),
		cosmosproto.E_AcceptsInterface,
	).(string)
	require.Equal(t, NFTDataInterfaceName, acceptedInterface)
}
