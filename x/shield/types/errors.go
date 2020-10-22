package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrNotShieldAdmin             = sdkerrors.Register(ModuleName, 101, "not the shield admin account")
	ErrNoDeposit                  = sdkerrors.Register(ModuleName, 102, "no coins given for initial deposit")
	ErrNoShield                   = sdkerrors.Register(ModuleName, 103, "no coins given for shield")
	ErrEmptySponsor               = sdkerrors.Register(ModuleName, 104, "no sponsor specified for a pool")
	ErrNoPoolFound                = sdkerrors.Register(ModuleName, 105, "no pool found")
	ErrNoUpdate                   = sdkerrors.Register(ModuleName, 106, "nothing was updated for the pool")
	ErrInvalidGenesis             = sdkerrors.Register(ModuleName, 107, "invalid genesis state")
	ErrInvalidPoolID              = sdkerrors.Register(ModuleName, 108, "invalid pool ID")
	ErrInvalidDuration            = sdkerrors.Register(ModuleName, 109, "invalid specification of coverage duration")
	ErrAdminWithdraw              = sdkerrors.Register(ModuleName, 110, "admin cannot manually withdraw collateral")
	ErrNoDelegationAmount         = sdkerrors.Register(ModuleName, 111, "cannot obtain delegation amount info")
	ErrInsufficientStaking        = sdkerrors.Register(ModuleName, 112, "insufficient total delegation amount to deposit the collateral")
	ErrPoolAlreadyPaused          = sdkerrors.Register(ModuleName, 113, "pool is already paused")
	ErrPoolAlreadyActive          = sdkerrors.Register(ModuleName, 114, "pool is already active")
	ErrPoolInactive               = sdkerrors.Register(ModuleName, 115, "pool is inactive")
	ErrPurchaseMissingDescription = sdkerrors.Register(ModuleName, 116, "missing description for the purchase")
	ErrNotEnoughShield            = sdkerrors.Register(ModuleName, 117, "not enough available shield")
	ErrNoPurchaseFound            = sdkerrors.Register(ModuleName, 118, "no purchase found for the given txhash")
	ErrNoRewards                  = sdkerrors.Register(ModuleName, 119, "no foreign coins rewards to transfer for the denomination")
	ErrInvalidDenom               = sdkerrors.Register(ModuleName, 120, "invalid coin denomination")
	ErrInvalidToAddr              = sdkerrors.Register(ModuleName, 121, "invalid recipient address")
	ErrNoCollateralFound          = sdkerrors.Register(ModuleName, 122, "no collateral for the pool found with the given provider address")
	ErrInvalidCollateralAmount    = sdkerrors.Register(ModuleName, 123, "invalid amount of collateral")
	ErrEmptySender                = sdkerrors.Register(ModuleName, 124, "no sender provided")
	ErrPoolLifeTooShort           = sdkerrors.Register(ModuleName, 125, "new pool life is too short")
	ErrPurchaseNotFound           = sdkerrors.Register(ModuleName, 126, "purchase is not found")
	ErrProviderNotFound           = sdkerrors.Register(ModuleName, 127, "provider is not found")
	ErrNotEnoughCollateral        = sdkerrors.Register(ModuleName, 128, "not enough collateral")
	ErrCompensationNotFound       = sdkerrors.Register(ModuleName, 129, "compensation is not found")
	ErrInvalidBeneficiary         = sdkerrors.Register(ModuleName, 130, "invalid beneficiary")
	ErrNotPayoutTime              = sdkerrors.Register(ModuleName, 131, "has not reached payout time yet")
	ErrOverWithdraw               = sdkerrors.Register(ModuleName, 132, "too much withdraw initiated")
	ErrNoPoolFoundForSponsor      = sdkerrors.Register(ModuleName, 133, "no pool found for the given sponsor")
	ErrSponsorAlreadyExists       = sdkerrors.Register(ModuleName, 134, "a pool already exists under the given sponsor")
	ErrCollateralBadDenom         = sdkerrors.Register(ModuleName, 135, "invalid coin denomination for collateral")
	ErrSponsorPurchase            = sdkerrors.Register(ModuleName, 136, "pool sponsor cannot purchase shield")
	ErrOperationNotSupported      = sdkerrors.Register(ModuleName, 137, "operation is currently not supported")
	ErrPoolShieldExceedsLimit     = sdkerrors.Register(ModuleName, 138, "pool shield exceeds limit")
	ErrShieldAdminNotActive       = sdkerrors.Register(ModuleName, 139, "shield admin is not activated")
	ErrPurchaseTooSmall           = sdkerrors.Register(ModuleName, 140, "the amount of purchased shield is too small")
)
