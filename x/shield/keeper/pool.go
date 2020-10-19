package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/certikfoundation/shentu/x/shield/types"
)

func (k Keeper) SetTotalCollateral(ctx sdk.Context, totalCollateral sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(totalCollateral)
	store.Set(types.GetTotalCollateralKey(), bz)
}

func (k Keeper) GetTotalCollateral(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTotalCollateralKey())
	if bz == nil {
		panic("total collateral is not found")
	}
	var totalCollateral sdk.Int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &totalCollateral)
	return totalCollateral
}

func (k Keeper) SetTotalShield(ctx sdk.Context, totalCollateral sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(totalCollateral)
	store.Set(types.GetTotalShieldKey(), bz)
}

func (k Keeper) GetTotalShield(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTotalShieldKey())
	if bz == nil {
		panic("total shield is not found")
	}
	var totalShield sdk.Int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &totalShield)
	return totalShield
}

func (k Keeper) SetTotalLocked(ctx sdk.Context, totalLocked sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(totalLocked)
	store.Set(types.GetTotalLockedKey(), bz)
}

func (k Keeper) GetTotalLocked(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTotalLockedKey())
	if bz == nil {
		panic("total shield is not found")
	}
	var totalLocked sdk.Int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &totalLocked)
	return totalLocked
}

func (k Keeper) SetServiceFees(ctx sdk.Context, totalCollateral types.MixedDecCoins) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(totalCollateral)
	store.Set(types.GetServiceFeesKey(), bz)
}

func (k Keeper) GetServiceFees(ctx sdk.Context) types.MixedDecCoins {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetServiceFeesKey())
	if bz == nil {
		panic("service fees is not found")
	}
	var serviceFees types.MixedDecCoins
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &serviceFees)
	return serviceFees
}

// SetPool sets data of a pool in kv-store.
func (k Keeper) SetPool(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(pool)
	store.Set(types.GetPoolKey(pool.ID), bz)
}

// GetPool gets data of a pool given pool ID.
func (k Keeper) GetPool(ctx sdk.Context, id uint64) (types.Pool, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolKey(id))
	if bz == nil {
		return types.Pool{}, false
	}
	var pool types.Pool
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &pool)
	return pool, true
}

// CreatePool creates a pool and sponsor's shield.
func (k Keeper) CreatePool(ctx sdk.Context, creator sdk.AccAddress, shield sdk.Coins, serviceFees types.MixedCoins, sponsor string, sponsorAddr sdk.AccAddress, protectionPeriod time.Duration, description string) (types.Pool, error) {
	admin := k.GetAdmin(ctx)
	if !creator.Equals(admin) {
		return types.Pool{}, types.ErrNotShieldAdmin
	}

	if _, found := k.GetPoolBySponsor(ctx, sponsor); found {
		return types.Pool{}, types.ErrSponsorAlreadyExists
	}

	// FIXME: This is incorrect. Should make sure the protection period is 21x days.
	if !k.ValidatePoolDuration(ctx, protectionPeriod) {
		return types.Pool{}, types.ErrPoolLifeTooShort
	}

	// Check available collaterals.
	shieldAmt := shield.AmountOf(k.sk.BondDenom(ctx))
	totalCollateral := k.GetTotalCollateral(ctx)
	totalShield := k.GetTotalShield(ctx)
	if totalShield.Add(shieldAmt).GT(totalCollateral) {
		return types.Pool{}, types.ErrNotEnoughCollateral
	}

	// Check pool shield limit.
	poolParams := k.GetPoolParams(ctx)
	maxShield := totalCollateral.ToDec().Mul(poolParams.PoolShieldLimit).TruncateInt()
	if shieldAmt.GT(maxShield) {
		return types.Pool{}, types.ErrPoolShieldExceedsLimit
	}

	// Transfer service fees to the Shield module account.
	if err := k.DepositNativeServiceFees(ctx, serviceFees.Native, creator); err != nil {
		return types.Pool{}, err
	}

	// Set the new project pool.
	id := k.GetNextPoolID(ctx)
	pool := types.NewPool(id, description, sponsor, sponsorAddr, shieldAmt)
	k.SetPool(ctx, pool)
	k.SetNextPoolID(ctx, id+1)

	// Update service fees in the global pool.
	serviceFeesUpdate := k.GetServiceFees(ctx)
	serviceFeesUpdate = serviceFeesUpdate.Add(types.MixedDecCoinsFromMixedCoins(serviceFees))
	k.SetServiceFees(ctx, serviceFeesUpdate)

	// Make a purchase for B.
	purchaseID := k.GetNextPurchaseID(ctx)
	protectionEndTime := ctx.BlockTime().Add(protectionPeriod)
	purchase := types.NewPurchase(purchaseID, protectionEndTime, "shield for sponsor", shieldAmt)
	k.InsertPurchaseQueue(ctx, types.NewPurchaseList(id, sponsorAddr, []types.Purchase{purchase}), protectionEndTime.Add(k.GetPurchaseDeletionPeriod(ctx)))
	k.AddPurchase(ctx, id, sponsorAddr, purchase)
	k.SetNextPurchaseID(ctx, purchaseID+1)

	return pool, nil
}

// FIXME UpdatePool only updates descriptions now. Any other things to be updated?
// UpdatePool updates pool info.
func (k Keeper) UpdatePool(ctx sdk.Context, poolID uint64, description string, updater sdk.AccAddress) (types.Pool, error) {
	admin := k.GetAdmin(ctx)
	if !updater.Equals(admin) {
		return types.Pool{}, types.ErrNotShieldAdmin
	}

	pool, found := k.GetPool(ctx, poolID)
	if !found {
		return types.Pool{}, types.ErrNoPoolFound
	}
	if description != "" {
		pool.Description = description
	}
	k.SetPool(ctx, pool)
	return pool, nil
}

// PausePool sets an active pool to be inactive.
func (k Keeper) PausePool(ctx sdk.Context, updater sdk.AccAddress, id uint64) (types.Pool, error) {
	admin := k.GetAdmin(ctx)
	if !updater.Equals(admin) {
		return types.Pool{}, types.ErrNotShieldAdmin
	}
	pool, found := k.GetPool(ctx, id)
	if !found {
		return types.Pool{}, types.ErrNoPoolFound
	}
	if !pool.Active {
		return types.Pool{}, types.ErrPoolAlreadyPaused
	}
	pool.Active = false
	k.SetPool(ctx, pool)
	return pool, nil
}

// ResumePool sets an inactive pool to be active.
func (k Keeper) ResumePool(ctx sdk.Context, updater sdk.AccAddress, id uint64) (types.Pool, error) {
	admin := k.GetAdmin(ctx)
	if !updater.Equals(admin) {
		return types.Pool{}, types.ErrNotShieldAdmin
	}
	pool, found := k.GetPool(ctx, id)
	if !found {
		return types.Pool{}, types.ErrNoPoolFound
	}
	if pool.Active {
		return types.Pool{}, types.ErrPoolAlreadyActive
	}
	pool.Active = true
	k.SetPool(ctx, pool)
	return pool, nil
}

// GetAllPools retrieves all pools in the store.
func (k Keeper) GetAllPools(ctx sdk.Context) (pools []types.Pool) {
	k.IterateAllPools(ctx, func(pool types.Pool) bool {
		pools = append(pools, pool)
		return false
	})
	return pools
}

// ClosePool closes the pool.
func (k Keeper) ClosePool(ctx sdk.Context, pool types.Pool) {
	// TODO: make sure nothing else needs to be done
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPoolKey(pool.ID))
}

// IterateAllPools iterates over the all the stored pools and performs a callback function.
func (k Keeper) IterateAllPools(ctx sdk.Context, callback func(pool types.Pool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.PoolKey)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pool types.Pool
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &pool)

		if callback(pool) {
			break
		}
	}
}

// ValidatePoolDuration validates new pool duration to be valid.
func (k Keeper) ValidatePoolDuration(ctx sdk.Context, timeDuration time.Duration) bool {
	poolParams := k.GetPoolParams(ctx)
	minPoolDuration := poolParams.MinPoolLife
	return timeDuration >= minPoolDuration
}
