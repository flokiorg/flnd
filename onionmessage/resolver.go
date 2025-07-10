package onionmessage

import (
	"context"
	"encoding/hex"

	graphdb "github.com/flokiorg/flnd/graph/db"
	"github.com/flokiorg/flnd/lnwire"
	"github.com/flokiorg/go-flokicoin/crypto"
)

type GraphNodeResolver struct {
	Graph  *graphdb.ChannelGraph
	OurPub *crypto.PublicKey
}

// RemotePubFromSCID resolves a node public key from a short channel ID.
func (r *GraphNodeResolver) RemotePubFromSCID(_ context.Context,
	scid lnwire.ShortChannelID) (*crypto.PublicKey, error) {

	log.Tracef("Resolving node public key for SCID %v", scid)

	edge, _, _, err := r.Graph.FetchChannelEdgesByID(scid.ToUint64())
	if err != nil {
		log.Debugf("Failed to fetch channel edges for SCID %v: %v",
			scid, err)

		return nil, err
	}

	otherNodeKeyBytes, err := edge.OtherNodeKeyBytes(
		r.OurPub.SerializeCompressed(),
	)
	if err != nil {
		log.Debugf("Failed to get other node key for SCID %v: %v",
			scid, err)

		return nil, err
	}

	pubKey, err := crypto.ParsePubKey(otherNodeKeyBytes[:])
	if err != nil {
		log.Debugf("Failed to parse public key for SCID %v: %v",
			scid, err)

		return nil, err
	}

	log.Tracef("Resolved SCID %v to node %s", scid,
		hex.EncodeToString(pubKey.SerializeCompressed()))

	return pubKey, nil
}
