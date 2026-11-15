package onionmessage

import (
	"testing"

	"github.com/flokiorg/go-flokicoin/crypto"
	"github.com/flokiorg/flnd/lnwire"
	"github.com/stretchr/testify/require"
)

func TestMockNodeIDResolverRemotePubFromSCID(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		resolver := newMockNodeIDResolver()
		priv, err := crypto.NewPrivateKey()
		require.NoError(t, err)
		pubKey := priv.PubKey()

		scid := lnwire.NewShortChanIDFromInt(1)
		resolver.addPeer(scid, pubKey)

		got, err := resolver.RemotePubFromSCID(t.Context(), scid)
		require.NoError(t, err)
		require.Equal(t, pubKey, got)
	})

	t.Run("unknown scid", func(t *testing.T) {
		t.Parallel()

		resolver := newMockNodeIDResolver()
		scid := lnwire.NewShortChanIDFromInt(2)

		got, err := resolver.RemotePubFromSCID(t.Context(), scid)
		require.Error(t, err)
		require.Nil(t, got)
	})
}
