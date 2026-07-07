package types

import (
	"encoding/json"
	"fmt"
	"io"

	"cosmossdk.io/x/tx/signing/aminojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	stringsTypeName                  = "panacea.did.v2.Strings"
	verificationRelationshipTypeName = "panacea.did.v2.VerificationRelationship"
)

// AminoJSONEncoderModifiers returns v0.50 aminojson encoders for DID types
// that previously used custom Go JSON marshalers.
func AminoJSONEncoderModifiers() []func(aminojson.Encoder) aminojson.Encoder {
	return []func(aminojson.Encoder) aminojson.Encoder{
		func(encoder aminojson.Encoder) aminojson.Encoder {
			return encoder.DefineTypeEncoding(stringsTypeName, encodeJSONStringOrStrings)
		},
		func(encoder aminojson.Encoder) aminojson.Encoder {
			return encoder.DefineTypeEncoding(verificationRelationshipTypeName, encodeVerificationRelationship)
		},
	}
}

func encodeJSONStringOrStrings(_ *aminojson.Encoder, msg protoreflect.Message, writer io.Writer) error {
	valuesField := msg.Descriptor().Fields().ByName("values")
	if valuesField == nil {
		return fmt.Errorf("field values not found in message %s", msg.Descriptor().FullName())
	}

	valuesList := msg.Get(valuesField).List()
	values := make([]string, valuesList.Len())
	for i := 0; i < valuesList.Len(); i++ {
		values[i] = valuesList.Get(i).String()
	}

	var value any = values
	if len(values) == 1 {
		value = values[0]
	}

	return jsonMarshal(writer, value)
}

func encodeVerificationRelationship(_ *aminojson.Encoder, msg protoreflect.Message, writer io.Writer) error {
	fields := msg.Descriptor().Fields()

	verificationMethodIDField := fields.ByName("verification_method_id")
	if verificationMethodIDField == nil {
		return fmt.Errorf("field verification_method_id not found in message %s", msg.Descriptor().FullName())
	}
	if msg.Has(verificationMethodIDField) {
		return jsonMarshal(writer, msg.Get(verificationMethodIDField).String())
	}

	verificationMethodField := fields.ByName("verification_method")
	if verificationMethodField == nil {
		return fmt.Errorf("field verification_method not found in message %s", msg.Descriptor().FullName())
	}
	if msg.Has(verificationMethodField) {
		return jsonMarshal(writer, verificationMethodJSON(msg.Get(verificationMethodField).Message()))
	}

	return jsonMarshal(writer, "")
}

func verificationMethodJSON(msg protoreflect.Message) map[string]string {
	jsonValue := make(map[string]string)
	setStringField(jsonValue, msg, "id", "id")
	setStringField(jsonValue, msg, "type", "type")
	setStringField(jsonValue, msg, "controller", "controller")
	setStringField(jsonValue, msg, "public_key_base58", "public_key_base58")
	return jsonValue
}

func setStringField(jsonValue map[string]string, msg protoreflect.Message, protoName protoreflect.Name, jsonName string) {
	field := msg.Descriptor().Fields().ByName(protoName)
	if field == nil || !msg.Has(field) {
		return
	}

	value := msg.Get(field).String()
	if value != "" {
		jsonValue[jsonName] = value
	}
}

func jsonMarshal(writer io.Writer, value any) error {
	bz, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = writer.Write(bz)
	return err
}
