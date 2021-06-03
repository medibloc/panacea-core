package types

const (
	// ModuleName defines the module name
	ModuleName = "aol"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_capability"

	// this line is used by starport scaffolding # ibc/keys/name
)

// this line is used by starport scaffolding # ibc/keys/port

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	TopicKey      = "Topic-value-"
	TopicCountKey = "Topic-count-"
)

const (
	WriterKey      = "Writer-value-"
	WriterCountKey = "Writer-count-"
)

const (
	RecordKey      = "Record-value-"
	RecordCountKey = "Record-count-"
)

const (
	OwnerKey      = "Owner-value-"
	OwnerCountKey = "Owner-count-"
)
