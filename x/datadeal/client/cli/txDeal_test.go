package cli

import (
	"encoding/base64"
	"testing"

	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/require"
)

func TestReadDataCertFile(t *testing.T) {
	testCert := makeTestCert()
	parsedDataCert, err := readDataCertFile("./testdata/data_certificate_file.json")
	require.NoError(t, err)

	require.Equal(t, parsedDataCert.GetSignature(), testCert.GetSignature())
	require.Equal(t, parsedDataCert.UnsignedCert.GetDealId(), testCert.UnsignedCert.GetDealId())
	require.Equal(t, parsedDataCert.UnsignedCert.GetDataHash(), testCert.UnsignedCert.GetDataHash())
	require.Equal(t, parsedDataCert.UnsignedCert.GetEncryptedDataUrl(), testCert.UnsignedCert.GetEncryptedDataUrl())
	require.Equal(t, parsedDataCert.UnsignedCert.GetOracleAddress(), testCert.UnsignedCert.GetOracleAddress())
	require.Equal(t, parsedDataCert.UnsignedCert.GetRequesterAddress(), testCert.UnsignedCert.GetRequesterAddress())
}

func makeTestCert() types.DataCert {

	decodeDataHash, _ := base64.StdEncoding.DecodeString("ZGF0YUhhc2g=")
	decodeURL, _ := base64.StdEncoding.DecodeString("ZW5jcnlwdGVkRGF0YVVSTA==")

	unsignedDataValidationCertificate := types.UnsignedDataValidationCertificate{
		DealId:           1,
		DataHash:         decodeDataHash,
		EncryptedDataUrl: decodeURL,
		OracleAddress:    "panacea1ugrau4qqr9446rpuj0srjrxspz02dd9nmlrjg3",
		RequesterAddress: "panacea1fpfugtgpzux8spqpe3kyqqpyy6rular2zlpusu",
	}

	decodeSig, _ := base64.StdEncoding.DecodeString("c2lnbmF0dXJl")
	dataCert := types.DataCert{
		UnsignedCert: &unsignedDataCert,
		Signature:    decodeSig,
	}

	return dataCert
}
