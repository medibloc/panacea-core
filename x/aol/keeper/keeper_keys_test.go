package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/types/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetLastElement(t *testing.T) {
	sdk.GetConfig().SetBech32PrefixForAccount(util.Bech32PrefixAccAddr, util.Bech32PrefixAccPub)

	owner, _ := sdk.AccAddressFromBech32("panacea17kmacx3czkdnhtfueqzzxk9xqzapj453f23m5a")
	topic := string([]byte{0x88, 0x00, 0x99})
	writer, _ := sdk.AccAddressFromBech32("panacea170vvyzwdgxrc5s5uhqhu6wdr7ppfyngfqsm6us")

	ownerKey := OwnerKey(owner)
	require.Equal(t, owner.Bytes(), getLastElement(ownerKey, OwnerKey(sdk.AccAddress{})))

	topicKey := TopicKey(owner, topic)
	require.Equal(t, []byte(topic), getLastElement(topicKey, TopicKey(owner, "")))

	writerKey := ACLWriterKey(owner, topic, writer)
	require.Equal(t, writer.Bytes(), getLastElement(writerKey, ACLWriterKey(owner, topic, sdk.AccAddress{})))
}
