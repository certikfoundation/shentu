package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Pool struct {
	PoolID           uint64
	Active           bool
	Description      string
	Sponsor          string
	Premium          MixedDecCoins
	StartBlockHeight int64
	TotalCollateral  sdk.Coins
	Available        sdk.Int
	Shield           sdk.Coins
	EndTime          int64
	EndBlockHeight   int64
}

func NewPool(
	shield sdk.Coins, deposit MixedDecCoins, sponsor string,
	endTime, startBlockHeight, endBlockHeight int64, id uint64) Pool {
	return Pool{
		Shield:           shield,
		Premium:          deposit,
		Sponsor:          sponsor,
		Active:           true,
		TotalCollateral:  shield,
		EndTime:          endTime,
		StartBlockHeight: startBlockHeight,
		EndBlockHeight:   endBlockHeight,
		PoolID:           id,
	}
}

type Collateral struct {
	PoolID            uint64
	Provider          sdk.AccAddress
	Amount            sdk.Coins
	Description       string
	LockedCollaterals []LockedCollateral
}

func NewCollateral(pool Pool, provider sdk.AccAddress, amount sdk.Coins) Collateral {
	return Collateral{
		PoolID:   pool.PoolID,
		Provider: provider,
		Amount:   amount,
	}
}

type MixedCoins struct {
	Native  sdk.Coins
	Foreign sdk.Coins
}

func (mc MixedCoins) Add(a MixedCoins) MixedCoins {
	native := mc.Native.Add(a.Native...)
	foreign := mc.Foreign.Add(a.Foreign...)
	return MixedCoins{
		Native:  native,
		Foreign: foreign,
	}
}

func (mc MixedCoins) String() string {
	return append(mc.Native, mc.Foreign...).String()
}

type MixedDecCoins struct {
	Native  sdk.DecCoins
	Foreign sdk.DecCoins
}

func InitMixedDecCoins() MixedDecCoins {
	return MixedDecCoins{
		Native:  sdk.DecCoins{},
		Foreign: sdk.DecCoins{},
	}
}

func NewMixedDecCoins(native, foreign sdk.DecCoins) MixedDecCoins {
	return MixedDecCoins{
		Native:  native,
		Foreign: foreign,
	}
}

// MixedDecCoinsFromMixedCoins converts MixedCoins to MixedDecCoins.
func MixedDecCoinsFromMixedCoins(mc MixedCoins) MixedDecCoins {
	return MixedDecCoins{
		Native:  sdk.NewDecCoinsFromCoins(mc.Native...),
		Foreign: sdk.NewDecCoinsFromCoins(mc.Foreign...),
	}
}

// Add adds two MixedDecCoins type coins together.
func (mdc MixedDecCoins) Add(a MixedDecCoins) MixedDecCoins {
	return MixedDecCoins{
		Native:  mdc.Native.Add(a.Native...),
		Foreign: mdc.Foreign.Add(a.Foreign...),
	}
}

// MulDec multiplies native and foreign coins by a decimal.
func (mdc MixedDecCoins) MulDec(d sdk.Dec) MixedDecCoins {
	return MixedDecCoins{
		Native:  mdc.Native.MulDec(d),
		Foreign: mdc.Foreign.MulDec(d),
	}
}

// QuoDec divides native and foreign coins by a decimal.
func (mdc MixedDecCoins) QuoDec(d sdk.Dec) MixedDecCoins {
	return MixedDecCoins{
		Native:  mdc.Native.QuoDec(d),
		Foreign: mdc.Foreign.QuoDec(d),
	}
}

func (mdc MixedDecCoins) String() string {
	return append(mdc.Native, mdc.Foreign...).String()
}

type PendingPayout struct {
	Amount sdk.Dec
	ToAddr string
}

type PendingPayouts []PendingPayout

func NewPendingPayouts(amount sdk.Dec, to string) PendingPayout {
	return PendingPayout{
		Amount: amount,
		ToAddr: to,
	}
}

// Provider tracks A or C's total delegation, total collateral,
// and rewards.
type Provider struct {
	DelegationBonded sdk.Coins
	Collateral       sdk.Coins
	TotalLocked      sdk.Coins
	Rewards          MixedDecCoins
}

func NewProvider() Provider {
	return Provider{
		DelegationBonded: sdk.Coins{},
		Collateral:       sdk.Coins{},
		TotalLocked:      sdk.Coins{},
	}
}

type Purchase struct {
	PoolID             uint64
	Shield             sdk.Coins
	StartBlockHeight   int64
	ProtectionEndTime  time.Time
	ClaimPeriodEndTime time.Time
	Description        string
	Purchaser          sdk.AccAddress
}

func NewPurchase(
	poolID uint64, shield sdk.Coins, startBlockHeight int64, protectionEndTime, claimPeriodEndTime time.Time,
	description string, purchaser sdk.AccAddress) Purchase {
	return Purchase{
		PoolID:             poolID,
		Shield:             shield,
		StartBlockHeight:   startBlockHeight,
		ProtectionEndTime:  protectionEndTime,
		ClaimPeriodEndTime: claimPeriodEndTime,
		Description:        description,
		Purchaser:          purchaser,
	}
}

// Withdrawal stores an ongoing withdrawal of pool collateral.
type Withdrawal struct {
	PoolID  uint64         `json:"pool_id" yaml:"pool_id"`
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Amount  sdk.Coins      `json:"amount" yaml:"amount"`
}

func NewWithdrawal(poolID uint64, addr sdk.AccAddress, amount sdk.Coins) Withdrawal {
	return Withdrawal{
		PoolID:  poolID,
		Address: addr,
		Amount:  amount,
	}
}