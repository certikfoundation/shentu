package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Pool struct {
	PoolID           uint64
	Community        []Collateral
	Shield           sdk.Coins
	Premium          MixedDecCoins
	CertiK           Collateral
	Sponsor          string
	StartBlockHeight int64
	Description      string
	Active           bool
	TotalCollateral  sdk.Coins
	EndTime          int64
	EndBlockHeight   int64
}

func NewPool(
	operator sdk.AccAddress, shield sdk.Coins, deposit MixedDecCoins, sponsor string,
	endTime, startBlockHeight, endBlockHeight int64, id uint64) Pool {
	return Pool{
		Shield:           shield,
		CertiK:           NewCollateral(operator, shield),
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
	PoolID      uint
	Provider    sdk.AccAddress
	Amount      sdk.Coins
	Earnings    *MixedDecCoins
	Description string
}

func NewCollateral(provider sdk.AccAddress, amount sdk.Coins) Collateral {
	mdc := InitMixedDecCoins()
	return Collateral{
		Provider: provider,
		Amount:   amount,
		Earnings: &mdc,
	}
}

// ForeignCoins separates sdk.Coins to shield foreign coins
type ForeignCoins sdk.Coins
type ForeignDecCoins sdk.DecCoins

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

// Participant tracks A or C's total delegation, total collateral,
// and rewards.
type Participant struct {
	TotalDelegation sdk.Coins
	TotalCollateral sdk.Coins
}

func NewParticipant() Participant {
	return Participant{
		TotalDelegation: sdk.Coins{},
		TotalCollateral: sdk.Coins{},
	}
}

type Purchase struct {
	PoolID                  uint64
	Shield                  sdk.Coins
	StartBlockHeight        int64
	ProtectionPeriodEndTime time.Time
	ClaimPeriodEndTime      time.Time
	Description             string
	Purchaser               sdk.AccAddress
}

func NewPurchase(
	poolID uint64, shield sdk.Coins, startBlockHeight int64, protectionPeriodEndTime, claimPeriodEndTime time.Time,
	description string, purchaser sdk.AccAddress) Purchase {
	return Purchase{
		PoolID:                  poolID,
		Shield:                  shield,
		StartBlockHeight:        startBlockHeight,
		ProtectionPeriodEndTime: protectionPeriodEndTime,
		ClaimPeriodEndTime:      claimPeriodEndTime,
		Description:             description,
		Purchaser:               purchaser,
	}
}
