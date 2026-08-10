package types

import (
	"testing"

	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestCustomGetSignersValidate(t *testing.T) {
	registry := newSigningRegistry(t)
	require.NoError(t, registry.SigningContext().Validate())
}

func TestMsgAddRecordCustomGetSigners(t *testing.T) {
	desc := msgAddRecordDescriptor(t)
	msg := dynamicpb.NewMessage(desc)

	writer := sdk.AccAddress(bytesOf(1)).String()
	feePayer := sdk.AccAddress(bytesOf(2)).String()

	msg.Set(desc.Fields().ByName("writer_address"), protoreflect.ValueOfString(writer))
	signers, err := getMsgAddRecordSigners(msg)
	require.NoError(t, err)
	require.Equal(t, [][]byte{sdk.MustAccAddressFromBech32(writer)}, signers)

	msg.Set(desc.Fields().ByName("fee_payer_address"), protoreflect.ValueOfString(feePayer))
	signers, err = getMsgAddRecordSigners(msg)
	require.NoError(t, err)
	require.Equal(t, [][]byte{sdk.MustAccAddressFromBech32(feePayer), sdk.MustAccAddressFromBech32(writer)}, signers)

	msg.Set(desc.Fields().ByName("fee_payer_address"), protoreflect.ValueOfString(writer))
	signers, err = getMsgAddRecordSigners(msg)
	require.NoError(t, err)
	require.Equal(t, [][]byte{sdk.MustAccAddressFromBech32(writer)}, signers)
}

func TestMsgAddRecordLegacyGetSignersDeduplicatesFeePayer(t *testing.T) {
	writer := sdk.AccAddress(bytesOf(1)).String()
	feePayer := sdk.AccAddress(bytesOf(2)).String()

	msg := MsgAddRecordRequest{
		WriterAddress:   writer,
		FeePayerAddress: feePayer,
	}
	require.Equal(t, []sdk.AccAddress{sdk.MustAccAddressFromBech32(feePayer), sdk.MustAccAddressFromBech32(writer)}, msg.GetSigners())

	msg.FeePayerAddress = writer
	require.Equal(t, []sdk.AccAddress{sdk.MustAccAddressFromBech32(writer)}, msg.GetSigners())
}

func newSigningRegistry(t *testing.T) types.InterfaceRegistry {
	t.Helper()

	signingOptions := txsigning.Options{
		AddressCodec:          authcodec.NewBech32Codec("panacea"),
		ValidatorAddressCodec: authcodec.NewBech32Codec("panaceavaloper"),
	}
	for _, signer := range CustomGetSigners() {
		signingOptions.DefineCustomGetSigners(signer.MsgType, signer.Fn)
	}

	registry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     gogoproto.HybridResolver,
		SigningOptions: signingOptions,
	})
	require.NoError(t, err)
	RegisterInterfaces(registry)

	return registry
}

func msgAddRecordDescriptor(t *testing.T) protoreflect.MessageDescriptor {
	t.Helper()

	desc, err := gogoproto.HybridResolver.FindDescriptorByName(msgAddRecordRequestName)
	require.NoError(t, err)

	msgDesc, ok := desc.(protoreflect.MessageDescriptor)
	require.True(t, ok)

	return msgDesc
}

func bytesOf(v byte) []byte {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = v
	}
	return bz
}
