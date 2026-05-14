package channeldb

import "github.com/flokiorg/flnd/chanstate"

// Compile-time assertions that ChannelStateDB satisfies the channel-state
// store contracts while the KV implementation still lives in channeldb.
var _ chanstate.Store[*OpenChannel] = (*ChannelStateDB)(nil)
