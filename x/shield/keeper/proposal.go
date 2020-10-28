package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/certikfoundation/shentu/x/shield/types"
)

// SecureCollaterals is called after a claim is submitted to secure
// the given amount of collaterals for the duration and adjust shield
// module states accordingly.
func (k Keeper) SecureCollaterals(ctx sdk.Context, poolID uint64, purchaser sdk.AccAddress, purchaseID uint64, loss sdk.Coins, duration time.Duration) error {
	lossAmt := loss.AmountOf(k.sk.BondDenom(ctx))

	// Verify shield.
	pool, found := k.GetPool(ctx, poolID)
	if !found {
		return types.ErrNoPoolFound
	}
	if lossAmt.GT(pool.Shield) {
		return types.ErrNotEnoughShield
	}

	// Verify collateral availability.
	totalCollateral := k.GetTotalCollateral(ctx)
	totalClaimed := k.GetTotalClaimed(ctx)
	totalClaimed = totalClaimed.Add(lossAmt)
	if totalClaimed.GT(totalCollateral) {
		panic("total claimed surpassed total collateral")
	}

	// Secure the updated loss ratio from each provider.
	providers := k.GetAllProviders(ctx)
	lossRatio := totalClaimed.ToDec().Quo(totalCollateral.ToDec())
	for i := range providers {
		lockAmt := providers[i].Collateral.ToDec().Mul(lossRatio).TruncateInt()
		if lockAmt.LT(providers[i].Collateral) {
			lockAmt = lockAmt.Add(sdk.OneInt())
		}
		k.SecureFromProvider(ctx, providers[i], lockAmt, duration)
	}

	// Update purchase states.
	purchaseList, found := k.GetPurchaseList(ctx, poolID, purchaser)
	if !found {
		return types.ErrPurchaseNotFound
	}
	var index int
	for i, entry := range purchaseList.Entries {
		if entry.PurchaseID == purchaseID {
			index = i
			break
		}
	}
	purchase := &purchaseList.Entries[index]
	if lossAmt.GT(purchase.Shield) {
		return types.ErrNotEnoughShield
	}
	k.DequeuePurchase(ctx, purchaseList, purchase.DeletionTime)
	purchase.Shield = purchase.Shield.Sub(lossAmt)
	votingEndTime := ctx.BlockTime().Add(duration)
	if purchase.DeletionTime.Before(votingEndTime) {
		// TODO: confirm this is correct
		purchase.DeletionTime = votingEndTime
	}
	k.SetPurchaseList(ctx, purchaseList)
	k.InsertExpiringPurchaseQueue(ctx, purchaseList, purchase.DeletionTime)

	// Update pool and global pool states.
	pool.Shield = pool.Shield.Sub(lossAmt)
	k.SetPool(ctx, pool)

	totalShield := k.GetTotalShield(ctx)
	totalShield = totalShield.Sub(lossAmt)
	k.SetTotalShield(ctx, totalShield)

	return nil
}

// SecureFromProvider secures the specified amount of collaterals from
// the provider for the duration. If necessary, it extends withdrawing
// collaterals and, if exist, their linked unbondings as well.
func (k Keeper) SecureFromProvider(ctx sdk.Context, provider types.Provider, amount sdk.Int, duration time.Duration) {
	// If there are enough bonded delegations backing
	// locked collaterals, we are done.
	if provider.DelegationBonded.GTE(amount) {
		k.SetProvider(ctx, provider.Address, provider)
		return
	}

	// Lenient check:
	// Check if non-withdrawing collaterals can cover the amount.
	if amount.GT(provider.Collateral.Sub(provider.Withdrawing)) {
		// Stricter check:
		// Consider the amount of all collaterals that would
		// remain deposited until the lock period ends.
		endTime := ctx.BlockTime().Add(duration)
		upcomingWithdrawAmount := k.ComputeWithdrawAmountByTime(ctx, provider.Address, endTime)
		availableCollateral := provider.Collateral.Sub(upcomingWithdrawAmount)
		if amount.GT(availableCollateral) {
			// Delay some withdrawals to cover the amount.
			delayAmt := amount.Sub(availableCollateral)
			k.DelayWithdraws(ctx, provider.Address, delayAmt, duration)
		}
	}
	k.SetProvider(ctx, provider.Address, provider)
}

func (k Keeper) RestoreShield(ctx sdk.Context, poolID uint64, purchaser sdk.AccAddress, id uint64, loss sdk.Coins) error {
	lossAmt := loss.AmountOf(k.sk.BondDenom(ctx))

	// Update the total shield.
	totalShield := k.GetTotalShield(ctx)
	totalShield = totalShield.Add(lossAmt)
	k.SetTotalShield(ctx, totalShield)

	// Update shield of the pool.
	pool, found := k.GetPool(ctx, poolID)
	if !found {
		return types.ErrNoPoolFound
	}
	pool.Shield = pool.Shield.Add(lossAmt)
	k.SetPool(ctx, pool)

	// Update shield of the purchase.
	purchaseList, found := k.GetPurchaseList(ctx, poolID, purchaser)
	if !found {
		return types.ErrPurchaseNotFound
	}
	for i := range purchaseList.Entries {
		if purchaseList.Entries[i].PurchaseID == id {
			purchaseList.Entries[i].Shield = purchaseList.Entries[i].Shield.Add(lossAmt)
			break
		}
	}
	k.SetPurchaseList(ctx, purchaseList)

	return nil
}
