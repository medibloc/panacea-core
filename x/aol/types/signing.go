package types

import (
	"fmt"

	txsigning "cosmossdk.io/x/tx/signing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const msgAddRecordRequestName = protoreflect.FullName("panacea.aol.v2.MsgAddRecordRequest")

// CustomGetSigners returns v0.50 signer hooks for messages whose signer set
// cannot be represented by static cosmos.msg.v1.signer annotations.
func CustomGetSigners() []txsigning.CustomGetSigner {
	return []txsigning.CustomGetSigner{
		{
			MsgType: msgAddRecordRequestName,
			Fn:      getMsgAddRecordSigners,
		},
	}
}

func getMsgAddRecordSigners(msg proto.Message) ([][]byte, error) {
	reflectMsg := msg.ProtoReflect()
	if reflectMsg.Descriptor().FullName() != msgAddRecordRequestName {
		return nil, fmt.Errorf("unexpected message type %s", reflectMsg.Descriptor().FullName())
	}

	writerAddress, err := stringField(reflectMsg, "writer_address")
	if err != nil {
		return nil, err
	}

	feePayerAddress, err := stringField(reflectMsg, "fee_payer_address")
	if err != nil {
		return nil, err
	}

	var signers [][]byte
	writer, err := sdk.AccAddressFromBech32(writerAddress)
	if err != nil {
		return nil, err
	}

	if feePayerAddress != "" {
		feePayer, err := sdk.AccAddressFromBech32(feePayerAddress)
		if err != nil {
			return nil, err
		}

		if !feePayer.Equals(writer) {
			signers = append(signers, feePayer)
		}
	}
	signers = append(signers, writer)

	return signers, nil
}

func stringField(msg protoreflect.Message, name protoreflect.Name) (string, error) {
	field := msg.Descriptor().Fields().ByName(name)
	if field == nil {
		return "", fmt.Errorf("field %s not found in message %s", name, msg.Descriptor().FullName())
	}
	return msg.Get(field).String(), nil
}
