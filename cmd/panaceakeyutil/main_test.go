package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
)

func TestTest(t *testing.T) {
	b := sha256.New().Sum([]byte("tendermint/PubKeySecp256k1"))
	var b2 [512]byte
	hex.Encode(b2[:], b)
	fmt.Println(hex.EncodeToString(b))
	fmt.Println(b2)
	cdc := amino.NewCodec()

	var a [33]byte
	a[0] = 1
	a[1] = 2
	a[2] = 3
	a[3] = 4
	cdc.RegisterConcrete(&a, "auth/StdTx", nil)
	mb, err := cdc.MarshalBinaryBare(a)
	require.NoError(t, err)
	t.Log(hex.EncodeToString(mb))
	jb, err := cdc.MarshalJSON(a)
	require.NoError(t, err)
	t.Log(string(jb))
}
