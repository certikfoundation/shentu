package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pool contains a shield project pool's data.
type Pool struct {
	// ID is the id of the pool.
	ID uint64 `json:"id" yaml:"id"`

	// Description is the term of the pool.
	Description string `json:"description" yaml:"description"`

	// Sponsor is the project owner of the pool.
	Sponsor string `json:"sponsor" yaml:"sponsor"`

	// SponsorAddress is the CertiK Chain address of the sponsor.
	SponsorAddress sdk.AccAddress `json:"sponsor_address" yaml:"sponsor_address"`

	// ShieldLimit is the maximum shield can be purchased for the pool.
	ShieldLimit sdk.Int `json:"shield_limit" yaml:"shield_limit"`

	// Active means new purchases are allowed.
	Active bool `json:"active" yaml:"active"`

	// Shield is the amount of all active purchased shields.
	Shield sdk.Int `json:"shield" yaml:"shield"`
}

// NewPool creates a new project pool.
func NewPool(id uint64, description, sponsor string, sponsorAddress sdk.AccAddress, shieldLimit sdk.Int, shield sdk.Int) Pool {
	return Pool{
		ID:             id,
		Description:    description,
		Sponsor:        sponsor,
		SponsorAddress: sponsorAddress,
		ShieldLimit:    shieldLimit,
		Active:         true,
		Shield:         shield,
	}
}

// Provider tracks total delegation, total collateral, and rewards of a provider.
type Provider struct {
	// Address is the address of the provider.
	Address sdk.AccAddress `json:"address" yaml:"address"`

	// DelegationBonded is the amount of bonded delegation.
	DelegationBonded sdk.Int `json:"delegation_bonded" yaml:"delegation_bonded"`

	// Collateral is amount of all collaterals for the provider, including
	// those in withdraw queue but excluding those currently locked, in all
	// pools.
	Collateral sdk.Int `json:"collateral" yaml:"collateral"`

	// Locked is the amount locked for pending claims.
	Locked sdk.Int `json:"total_locked" yaml:"total_locked"`

	// LockedCollaterals are collaterals locked for different proposals.
	LockedCollaterals []LockedCollateral `json:"locked_collaterals" yaml:"locked_collaterals"`

	// Withdrawing is the amount of collateral in withdraw queues.
	Withdrawing sdk.Int `json:"withdrawing" yaml:"withdrawing"`

	// Rewards is the pooling rewards to be collected.
	Rewards MixedDecCoins `json:"rewards" yaml:"rewards"`
}

// NewProvider creates a new provider object.
func NewProvider(addr sdk.AccAddress) Provider {
	return Provider{
		Address:          addr,
		DelegationBonded: sdk.ZeroInt(),
		Collateral:       sdk.ZeroInt(),
		Locked:           sdk.ZeroInt(),
		Withdrawing:      sdk.ZeroInt(),
	}
}

// LockedCollateral defines the data type of locked collateral for a claim proposal.
type LockedCollateral struct {
	ProposalID uint64  `json:"proposal_id" yaml:"proposal_id"`
	Amount     sdk.Int `json:"locked_coins" yaml:"locked_coins"`
}

// NewLockedCollateral returns a new LockedCollateral instance.
func NewLockedCollateral(proposalID uint64, lockedAmt sdk.Int) LockedCollateral {
	return LockedCollateral{
		ProposalID: proposalID,
		Amount:     lockedAmt,
	}
}

// Purchase record an individual purchase.
type Purchase struct {
	// PurchaseID is the purchase_id.
	PurchaseID uint64 `json:"purchase_id" yaml:"purchase_id"`

	// ProtectionEndTime is the time when the protection of the shield ends.
	ProtectionEndTime time.Time `json:"protection_end_time" yaml:"protection_end_time"`

	// DeletionTime is the time when the purchase should be deleted.
	DeletionTime time.Time `json:"deletion_time" yaml:"deletion_time"`

	// Description is the information about the protected asset.
	Description string `json:"description" yaml:"description"`

	// Shield is the unused amount of shield purchased.
	Shield sdk.Int `json:"shield" yaml:"shield"`

	// ServiceFees is the service fees paid by this purchase.
	ServiceFees MixedDecCoins `json:"service_fees" yaml:"service_fees"`
}

// NewPurchase creates a new purchase object.
func NewPurchase(purchaseID uint64, protectionEndTime, deletionTime time.Time, description string, shield sdk.Int, serviceFees MixedDecCoins) Purchase {
	return Purchase{
		PurchaseID:        purchaseID,
		ProtectionEndTime: protectionEndTime,
		DeletionTime:      deletionTime,
		Description:       description,
		Shield:            shield,
		ServiceFees:       serviceFees,
	}
}

// PurchaseList is a collection of purchase.
type PurchaseList struct {
	// PoolID is the id of the shield of the purchase.
	PoolID uint64 `json:"pool_id" yaml:"pool_id"`

	// Purchaser is the address making the purchase.
	Purchaser sdk.AccAddress `json:"purchaser" yaml:"purchaser"`

	// Entries stores all purchases by the purchaser in the pool.
	Entries []Purchase `json:"entries" yaml:"entries"`
}

// NewPurchaseList creates a new purchase list.
func NewPurchaseList(poolID uint64, purchaser sdk.AccAddress, purchases []Purchase) PurchaseList {
	return PurchaseList{
		PoolID:    poolID,
		Purchaser: purchaser,
		Entries:   purchases,
	}
}

// PoolPurchase is a pair of pool id and purchaser.
type PoolPurchaser struct {
	// PoolID is the id of the shield pool.
	PoolID uint64

	// Purchaser is the chain address of the purchaser.
	Purchaser sdk.AccAddress
}

// Withdraw stores an ongoing withdraw of pool collateral.
type Withdraw struct {
	// Address is the chain address of the provider withdrawing.
	Address sdk.AccAddress `json:"address" yaml:"address"`

	// Amount is the amount of withdraw.
	Amount sdk.Int `json:"amount" yaml:"amount"`

	// CompletionTime is the scheduled withdraw completion time.
	CompletionTime time.Time `json:"completion_time" yaml:"completion_time"`

	// LinkedUnbonding stores information about the unbonding that
	// triggered the withdraw, which may or may not exist.
	LinkedUnbonding UnbondingInfo `json:"linked_unbonding" yaml:"linked_unbonding"`
}

// NewWithdraw creates a new withdraw object.
func NewWithdraw(addr sdk.AccAddress, amount sdk.Int, completionTime time.Time, ubdInfo UnbondingInfo) Withdraw {
	return Withdraw{
		Address:         addr,
		Amount:          amount,
		CompletionTime:  completionTime,
		LinkedUnbonding: ubdInfo,
	}
}

// Withdraws contains multiple withdraws.
type Withdraws []Withdraw

type UnbondingInfo struct {
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	CompletionTime   time.Time      `json:"completion_time" yaml:"completion_time"`
	Confirmed        bool           `json:"confirmed" yaml:"confirmed"`
}

func NewUnbondingInfo(valAddr sdk.ValAddress, completionTime time.Time, confirmed bool) UnbondingInfo {
	return UnbondingInfo{
		ValidatorAddress: valAddr,
		CompletionTime:   completionTime,
		Confirmed:        confirmed,
	}
}
