package types_test

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/x/tx/signing/aminojson"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/medibloc/panacea-core/v2/x/did/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestAminoJSONEncoderStringsMatchesLegacyJSON(t *testing.T) {
	encoder := newDIDAminoJSONEncoder()

	testCases := []struct {
		name   string
		values []string
	}{
		{
			name:   "empty",
			values: []string{},
		},
		{
			name:   "single",
			values: []string{types.ContextDIDV1},
		},
		{
			name:   "multiple",
			values: []string{types.ContextDIDV1, "https://example.com/did/context/v1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := newDynamicMessage(t, "panacea.did.v2.Strings")
			valuesField := msg.Descriptor().Fields().ByName("values")
			require.NotNil(t, valuesField)

			valuesList := msg.Mutable(valuesField).List()
			for _, value := range tc.values {
				valuesList.Append(protoreflect.ValueOfString(value))
			}

			actual, err := encoder.Marshal(msg)
			require.NoError(t, err)

			legacyValue := types.JSONStringOrStrings(tc.values)
			legacyJSON, err := legacyValue.MarshalJSON()
			require.NoError(t, err)
			require.Equal(t, string(sortedJSON(t, legacyJSON)), string(actual))
		})
	}
}

func TestAminoJSONEncoderVerificationRelationshipMatchesLegacyJSON(t *testing.T) {
	encoder := newDIDAminoJSONEncoder()
	did, verificationMethodID, verificationMethod := testVerificationMethod()

	testCases := []struct {
		name     string
		setup    func(*dynamicpb.Message)
		legacy   types.VerificationRelationship
		validate func(types.VerificationRelationship)
	}{
		{
			name:   "empty",
			setup:  func(*dynamicpb.Message) {},
			legacy: types.VerificationRelationship{},
		},
		{
			name: "verification method id",
			setup: func(msg *dynamicpb.Message) {
				field := msg.Descriptor().Fields().ByName("verification_method_id")
				require.NotNil(t, field)
				msg.Set(field, protoreflect.ValueOfString(verificationMethodID))
			},
			legacy: types.NewVerificationRelationship(verificationMethodID),
			validate: func(rel types.VerificationRelationship) {
				require.True(t, rel.Valid(did))
			},
		},
		{
			name: "embedded verification method",
			setup: func(msg *dynamicpb.Message) {
				field := msg.Descriptor().Fields().ByName("verification_method")
				require.NotNil(t, field)
				methodMsg := msg.Mutable(field).Message()
				setDynamicStringField(t, methodMsg, "id", verificationMethod.Id)
				setDynamicStringField(t, methodMsg, "type", verificationMethod.Type)
				setDynamicStringField(t, methodMsg, "controller", verificationMethod.Controller)
				setDynamicStringField(t, methodMsg, "public_key_base58", verificationMethod.PublicKeyBase58)
			},
			legacy: types.NewVerificationRelationshipDedicated(verificationMethod),
			validate: func(rel types.VerificationRelationship) {
				require.True(t, rel.Valid(did))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := newDynamicMessage(t, "panacea.did.v2.VerificationRelationship")
			tc.setup(msg)

			actual, err := encoder.Marshal(msg)
			require.NoError(t, err)

			legacyJSON, err := tc.legacy.MarshalJSON()
			require.NoError(t, err)
			require.Equal(t, string(sortedJSON(t, legacyJSON)), string(actual))
			if tc.validate != nil {
				tc.validate(tc.legacy)
			}
		})
	}
}

func TestVerificationRelationshipUnmarshalRejectsOneofWrapperJSON(t *testing.T) {
	_, verificationMethodID, verificationMethod := testVerificationMethod()

	var relationship types.VerificationRelationship
	err := relationship.UnmarshalJSON([]byte(`{
		"verification_method_id": "` + verificationMethodID + `",
		"verification_method": {
			"id": "` + verificationMethod.Id + `",
			"type": "` + verificationMethod.Type + `",
			"controller": "` + verificationMethod.Controller + `",
			"publicKeyBase58": "` + verificationMethod.PublicKeyBase58 + `"
		}
	}`))
	require.Error(t, err)
}

func TestVerificationRelationshipBinaryUnmarshalDuplicateOneofUsesLastValue(t *testing.T) {
	_, verificationMethodID, verificationMethod := testVerificationMethod()
	idRelationship := types.NewVerificationRelationship(verificationMethodID)
	methodRelationship := types.NewVerificationRelationshipDedicated(verificationMethod)

	idBytes, err := idRelationship.Marshal()
	require.NoError(t, err)
	methodBytes, err := methodRelationship.Marshal()
	require.NoError(t, err)

	var relationship types.VerificationRelationship
	require.NoError(t, relationship.Unmarshal(append(append([]byte{}, idBytes...), methodBytes...)))
	require.Equal(t, methodRelationship, relationship)

	require.NoError(t, relationship.Unmarshal(append(append([]byte{}, methodBytes...), idBytes...)))
	require.Equal(t, idRelationship, relationship)
}

func newDIDAminoJSONEncoder() aminojson.Encoder {
	encoder := aminojson.NewEncoder(aminojson.EncoderOptions{})
	for _, modifier := range types.AminoJSONEncoderModifiers() {
		encoder = modifier(encoder)
	}
	return encoder
}

func newDynamicMessage(t *testing.T, typeName protoreflect.FullName) *dynamicpb.Message {
	t.Helper()

	desc, err := gogoproto.HybridResolver.FindDescriptorByName(typeName)
	require.NoError(t, err)
	msgDesc, ok := desc.(protoreflect.MessageDescriptor)
	require.True(t, ok)

	return dynamicpb.NewMessage(msgDesc)
}

func setDynamicStringField(t *testing.T, msg protoreflect.Message, fieldName protoreflect.Name, value string) {
	t.Helper()

	field := msg.Descriptor().Fields().ByName(fieldName)
	require.NotNil(t, field)
	msg.Set(field, protoreflect.ValueOfString(value))
}

func sortedJSON(t *testing.T, bz []byte) []byte {
	t.Helper()

	var value any
	require.NoError(t, json.Unmarshal(bz, &value))
	sorted, err := json.Marshal(value)
	require.NoError(t, err)
	return sorted
}

func testVerificationMethod() (string, string, types.VerificationMethod) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, []byte{1, 2, 3, 4, 5})
	return did, verificationMethodID, verificationMethod
}
