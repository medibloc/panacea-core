package cli

import (
	"encoding/base64"
	"testing"

	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/require"
)

func TestParseSellDataFlags(t *testing.T) {
	testCert := makeTestCert()
	parsedDataCert, err := readDataCertFile("./testdata/data_certificate_file.json")
	require.NoError(t, err)

	require.Equal(t, parsedDataCert.GetSignature(), testCert.GetSignature())
	require.Equal(t, parsedDataCert.UnsignedCert.GetDealId(), testCert.UnsignedCert.GetDealId())
	require.Equal(t, parsedDataCert.UnsignedCert.GetDataHash(), testCert.UnsignedCert.GetDataHash())
	require.Equal(t, parsedDataCert.UnsignedCert.GetEncryptedDataUrl(), testCert.UnsignedCert.GetEncryptedDataUrl())
	require.Equal(t, parsedDataCert.UnsignedCert.GetDataValidatorAddress(), testCert.UnsignedCert.GetDataValidatorAddress())
	require.Equal(t, parsedDataCert.UnsignedCert.GetRequesterAddress(), testCert.UnsignedCert.GetRequesterAddress())
}

func makeTestCert() types.DataValidationCertificate {

	decodeDataHash, _ := base64.StdEncoding.DecodeString("ZGF0YUhhc2g=")
	decodeURL, _ := base64.StdEncoding.DecodeString("ZW5jcnlwdGVkRGF0YVVSTA==")

	unsignedDataValidationCertificate := types.UnsignedDataValidationCertificate{
		DealId:               1,
		DataHash:             decodeDataHash,
		EncryptedDataUrl:     decodeURL,
		DataValidatorAddress: "panacea1ugrau4qqr9446rpuj0srjrxspz02dd9nmlrjg3",
		RequesterAddress:     "panacea1fpfugtgpzux8spqpe3kyqqpyy6rular2zlpusu",
	}

	decodeSig, _ := base64.StdEncoding.DecodeString("c2lnbmF0dXJl")
	dataValidationCertificate := types.DataValidationCertificate{
		UnsignedCert: &unsignedDataValidationCertificate,
		Signature:    decodeSig,
	}

	return dataValidationCertificate
}
