package types

import (
	"encoding/binary"
)

const (
	// ModuleName is the name of this module
	ModuleName = "shield"

	// RouterKey is used to route messages.
	RouterKey = ModuleName

	// StoreKey is the prefix under which we store this module's data.
	StoreKey = ModuleName

	// QuerierRoute is used to handle abci_query requests.
	QuerierRoute = ModuleName

	// DefaultParamspace is the default name for parameter store.
	DefaultParamspace = ModuleName
)

var (
	PoolKey           = []byte{0x0}
	ShieldOperatorKey = []byte{0x1}
	NextPoolIDKey     = []byte{0x2}
)

// GetPoolKey gets the key for the pool identified by pool ID.
func GetPoolKey(id uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	return append(PoolKey, b...)
}

// GetShieldOperatorKey gets the key for the shield admin.
func GetShieldOperatorKey() []byte {
	return ShieldOperatorKey
}

// GetNextPoolIDKey gets the key for the latest pool ID.
func GetNextPoolIDKey() []byte {
	return NextPoolIDKey
}
