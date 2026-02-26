package onionmessage

import "errors"

var (
	// ErrNoPathFound is returned when no path exists between the source
	// and destination nodes that supports onion messaging.
	ErrNoPathFound = errors.New("no path found to destination")

	// ErrDestinationNoOnionSupport is returned when the destination node
	// does not advertise support for onion messages.
	ErrDestinationNoOnionSupport = errors.New("destination does not " +
		"support onion messages")

	// ErrNodeNotFound is returned when the node is not found in the graph.
	ErrNodeNotFound = errors.New("node not found in graph")
)
