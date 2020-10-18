package types

import (
	"encoding/binary"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of this module.
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
	ShieldAdminKey    = []byte{0x0}
	GlobalPoolKey     = []byte{0x1}
	PoolKey           = []byte{0x2}
	NextPoolIDKey     = []byte{0x3}
	NextPurchaseIDKey = []byte{0x4}
	PurchaseListKey   = []byte{0x5}
	PurchaseQueueKey  = []byte{0x6}
	ReimbursementKey  = []byte{0x7}
	CollateralKey     = []byte{0x8}
	ProviderKey       = []byte{0x9}
	WithdrawQueueKey  = []byte{0xA}
)

// GetGlobalPoolKey gets the key for the shield global pool.
func GetGlobalPoolKey() []byte {
	return GlobalPoolKey
}

// GetPoolKey gets the key for the pool identified by pool ID.
func GetPoolKey(id uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	return append(PoolKey, b...)
}

// GetShieldAdminKey gets the key for the shield admin.
func GetShieldAdminKey() []byte {
	return ShieldAdminKey
}

// GetNextPoolIDKey gets the key for the next pool ID.
func GetNextPoolIDKey() []byte {
	return NextPoolIDKey
}

// GetNextPurchaseIDKey gets the key for the next pool ID.
func GetNextPurchaseIDKey() []byte {
	return NextPurchaseIDKey
}

// GetPurchaseTxHashKey gets the key for a purchase.
func GetPurchaseListKey(id uint64, purchaser sdk.AccAddress) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, id)
	return append(PurchaseListKey, append(bz, purchaser.Bytes()...)...)
}

// GetProviderKey gets the key for the delegator's tracker.
func GetProviderKey(addr sdk.AccAddress) []byte {
	return append(ProviderKey, addr...)
}

// GetWithdrawCompletionTimeKey gets a withdraw queue key,
// which is obtained from the completion time.
func GetWithdrawCompletionTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(WithdrawQueueKey, bz...)
}

// GetPurchaseCompletionTimeKey gets a withdraw queue key,
// which is obtained from the completion time.
func GetPurchaseCompletionTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(PurchaseQueueKey, bz...)
}

// GetReimbursement gets the key for a reimbursement.
func GetReimbursementKey(proposalID uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(ReimbursementKey, bz...)
}

// GetCollateralKey gets the key for a collateral.
func GetCollateralKey(poolID uint64, address sdk.AccAddress) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, poolID)
	return append(CollateralKey, append(bz, address...)...)
}

// GetPoolCollateralsKey gets the key for collaterals of a pool.
func GetPoolCollateralsKey(poolID uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, poolID)
	return append(CollateralKey, bz...)
}
