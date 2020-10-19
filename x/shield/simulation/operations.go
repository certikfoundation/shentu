package simulation

import (
	"math/rand"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/certikfoundation/shentu/x/shield/keeper"
	"github.com/certikfoundation/shentu/x/shield/types"
)

const (
	// C's operations
	OpWeightMsgCreatePool   = "op_weight_msg_create_pool"
	OpWeightMsgUpdatePool   = "op_weight_msg_update_pool"
	OpWeightMsgClearPayouts = "op_weight_msg_clear_payouts"

	// B and C's operations
	OpWeightMsgDepositCollateral      = "op_weight_msg_deposit_collateral"
	OpWeightMsgWithdrawCollateral     = "op_weight_msg_withdraw_collateral"
	OpWeightMsgWithdrawRewards        = "op_weight_msg_withdraw_rewards"
	OpWeightMsgWithdrawForeignRewards = "op_weight_msg_withdraw_foreign_rewards"

	// P's operations
	OpWeightMsgPurchaseShield   = "op_weight_msg_purchase_shield"
	OpWeightShieldClaimProposal = "op_weight_msg_submit_claim_proposal"
)

var (
	DefaultWeightMsgCreatePool             = 10
	DefaultWeightMsgUpdatePool             = 10
	DefaultWeightMsgDepositCollateral      = 20
	DefaultWeightMsgWithdrawCollateral     = 20
	DefaultWeightMsgWithdrawRewards        = 10
	DefaultWeightMsgWithdrawForeignRewards = 10
	DefaultWeightMsgPurchaseShield         = 0
	DefaultWeightShieldClaimProposal       = 0

	DefaultIntMax = 100000000000
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, k keeper.Keeper, ak types.AccountKeeper, sk types.StakingKeeper) simulation.WeightedOperations {
	var weightMsgCreatePool int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePool, &weightMsgCreatePool, nil,
		func(_ *rand.Rand) {
			weightMsgCreatePool = DefaultWeightMsgCreatePool
		})
	var weightMsgUpdatePool int
	appParams.GetOrGenerate(cdc, OpWeightMsgUpdatePool, &weightMsgUpdatePool, nil,
		func(_ *rand.Rand) {
			weightMsgUpdatePool = DefaultWeightMsgUpdatePool
		})
	var weightMsgDepositCollateral int
	appParams.GetOrGenerate(cdc, OpWeightMsgDepositCollateral, &weightMsgDepositCollateral, nil,
		func(_ *rand.Rand) {
			weightMsgDepositCollateral = DefaultWeightMsgDepositCollateral
		})
	var weightMsgWithdrawCollateral int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawCollateral, &weightMsgWithdrawCollateral, nil,
		func(_ *rand.Rand) {
			weightMsgWithdrawCollateral = DefaultWeightMsgWithdrawCollateral
		})
	var weightMsgWithdrawRewards int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawRewards, &weightMsgWithdrawRewards, nil,
		func(_ *rand.Rand) {
			weightMsgWithdrawRewards = DefaultWeightMsgWithdrawRewards
		})
	var weightMsgWithdrawForeignRewards int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawForeignRewards, &weightMsgWithdrawForeignRewards, nil,
		func(_ *rand.Rand) {
			weightMsgWithdrawForeignRewards = DefaultWeightMsgWithdrawForeignRewards
		})
	var weightMsgPurchaseShield int
	appParams.GetOrGenerate(cdc, OpWeightMsgPurchaseShield, &weightMsgPurchaseShield, nil,
		func(_ *rand.Rand) {
			weightMsgPurchaseShield = DefaultWeightMsgPurchaseShield
		})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(weightMsgCreatePool, SimulateMsgCreatePool(k, ak, sk)),
		simulation.NewWeightedOperation(weightMsgUpdatePool, SimulateMsgUpdatePool(k, ak, sk)),
		simulation.NewWeightedOperation(weightMsgDepositCollateral, SimulateMsgDepositCollateral(k, ak, sk)),
		simulation.NewWeightedOperation(weightMsgWithdrawCollateral, SimulateMsgWithdrawCollateral(k, ak, sk)),
		simulation.NewWeightedOperation(weightMsgWithdrawRewards, SimulateMsgWithdrawRewards(k, ak, sk)),
		simulation.NewWeightedOperation(weightMsgWithdrawForeignRewards, SimulateMsgWithdrawForeignRewards(k, ak, sk)),
		//simulation.NewWeightedOperation(weightMsgPurchaseShield, SimulateMsgPurchaseShield(k, ak, sk)),
	}
}

// SimulateMsgCreatePool generates a MsgCreatePool object with all of its fields randomized.
func SimulateMsgCreatePool(k keeper.Keeper, ak types.AccountKeeper, sk types.StakingKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		pools := k.GetAllPools(ctx)
		// restrict number of pools to reduce gas consumptions for unbondings and redelegations
		if len(pools) > 20 {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		// admin
		var (
			adminAddr  sdk.AccAddress
			available  sdk.Int
			found      bool
			simAccount simulation.Account
		)

		adminAddr, available, found = keeper.RandomDelegation(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		if available.LT(sdk.OneInt()) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		k.SetAdmin(ctx, adminAddr)

		for _, simAcc := range accs {
			if simAcc.Address.Equals(adminAddr) {
				simAccount = simAcc
				break
			}
		}
		account := ak.GetAccount(ctx, simAccount.Address)
		bondDenom := sk.BondDenom(ctx)

		// shield
		provider, found := k.GetProvider(ctx, simAccount.Address)
		var shieldAmount sdk.Int
		var err error
		if found {
			shieldAmount, err = simulation.RandPositiveInt(r, provider.Available)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, nil
			}
		} else {
			shieldAmount, err = simulation.RandPositiveInt(r, available)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, nil
			}
		}
		shield := sdk.NewCoins(sdk.NewCoin(bondDenom, shieldAmount))

		// sponsor
		sponsor := strings.ToLower(simulation.RandStringOfLength(r, 10))
		_, found = k.GetPoolBySponsor(ctx, sponsor)
		if found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		// deposit
		nativeAmount := account.SpendableCoins(ctx.BlockTime()).AmountOf(bondDenom)
		if !nativeAmount.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		nativeAmount, err = simulation.RandPositiveInt(r, nativeAmount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		nativeDeposit := sdk.NewCoins(sdk.NewCoin(bondDenom, nativeAmount))
		foreignAmount, err := simulation.RandPositiveInt(r, sdk.NewInt(int64(DefaultIntMax)))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		foreignDeposit := sdk.NewCoins(sdk.NewCoin(sponsor, foreignAmount))
		deposit := types.MixedCoins{Native: nativeDeposit, Foreign: foreignDeposit}

		// time of coverage
		poolParams := k.GetPoolParams(ctx)
		minPoolLife := int(poolParams.MinPoolLife)

		timeOfCoverage := int64(simulation.RandIntBetween(r, minPoolLife, minPoolLife*10))
		coverageDuration := time.Duration(timeOfCoverage)
		sponsorAcc, _ := simulation.RandomAcc(r, accs)

		description := simulation.RandStringOfLength(r, 42)

		msg := types.NewMsgCreatePool(simAccount.Address, shield, deposit, sponsor, sponsorAcc.Address, coverageDuration, description)
		fees := sdk.Coins{}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		if _, _, err := app.Deliver(tx); err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgUpdatePool generates a MsgUpdatePool object with all of its fields randomized.
func SimulateMsgUpdatePool(k keeper.Keeper, ak types.AccountKeeper, sk types.StakingKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		adminAddr := k.GetAdmin(ctx)
		var simAccount simulation.Account
		for _, simAcc := range accs {
			if simAcc.Address.Equals(adminAddr) {
				simAccount = simAcc
				break
			}
		}
		account := ak.GetAccount(ctx, simAccount.Address)
		bondDenom := sk.BondDenom(ctx)

		// poolID and sponsor
		poolID, sponsor, found := keeper.RandomPoolInfo(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		// shield
		provider, found := k.GetProvider(ctx, adminAddr)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		shieldAmount, err := simulation.RandPositiveInt(r, provider.Available.Quo(sdk.NewInt(2)))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		shield := sdk.NewCoins(sdk.NewCoin(bondDenom, shieldAmount))

		// deposit
		nativeAmount := account.SpendableCoins(ctx.BlockTime()).AmountOf(bondDenom)
		if !nativeAmount.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		nativeAmount, err = simulation.RandPositiveInt(r, nativeAmount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		nativeDeposit := sdk.NewCoins(sdk.NewCoin(bondDenom, nativeAmount))
		foreignAmount, err := simulation.RandPositiveInt(r, sdk.NewInt(int64(DefaultIntMax)))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		foreignDeposit := sdk.NewCoins(sdk.NewCoin(sponsor, foreignAmount))
		deposit := types.MixedCoins{Native: nativeDeposit, Foreign: foreignDeposit}

		// time of coverage
		poolParams := k.GetPoolParams(ctx)
		minPoolLife := int(poolParams.MinPoolLife)

		timeOfCoverage := int64(simulation.RandIntBetween(r, minPoolLife, minPoolLife*10))
		coverageDuration := time.Duration(timeOfCoverage)

		msg := types.NewMsgUpdatePool(simAccount.Address, shield, deposit, poolID, coverageDuration, "")

		fees := sdk.Coins{}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		if _, _, err := app.Deliver(tx); err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgDepositCollateral generates a MsgDepositCollateral object with all of its fields randomized.
func SimulateMsgDepositCollateral(k keeper.Keeper, ak types.AccountKeeper, sk types.StakingKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		delAddr, delAmount, found := keeper.RandomDelegation(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		var simAccount simulation.Account
		for _, simAcc := range accs {
			if simAcc.Address.Equals(delAddr) {
				simAccount = simAcc
				break
			}
		}
		account := ak.GetAccount(ctx, simAccount.Address)

		// collateral coins
		provider, found := k.GetProvider(ctx, simAccount.Address)
		if found {
			delAmount = provider.Available
		}
		collateralAmount, err := simulation.RandPositiveInt(r, delAmount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		collateral := sdk.NewCoin(sk.BondDenom(ctx), collateralAmount)

		msg := types.NewMsgDepositCollateral(simAccount.Address, collateral)

		fees := sdk.Coins{}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		if _, _, err := app.Deliver(tx); err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgWithdrawCollateral generates a MsgWithdrawCollateral object with all of its fields randomized.
func SimulateMsgWithdrawCollateral(k keeper.Keeper, ak types.AccountKeeper, sk types.StakingKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		provider, found := keeper.RandomProvider(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		var simAccount simulation.Account
		for _, simAcc := range accs {
			if simAcc.Address.Equals(provider.Address) {
				simAccount = simAcc
				break
			}
		}
		account := ak.GetAccount(ctx, simAccount.Address)

		// withdraw coins
		withdrawable := provider.Collateral.Sub(provider.Withdrawing)
		withdrawAmount, err := simulation.RandPositiveInt(r, withdrawable)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		withdraw := sdk.NewCoin(sk.BondDenom(ctx), withdrawAmount)

		msg := types.NewMsgWithdrawCollateral(simAccount.Address, withdraw)

		fees := sdk.Coins{}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		if _, _, err := app.Deliver(tx); err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgWithdrawRewards generates a MsgWithdrawRewards object with all of its fields randomized.
func SimulateMsgWithdrawRewards(k keeper.Keeper, ak types.AccountKeeper, sk types.StakingKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		provider, found := keeper.RandomProvider(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		var simAccount simulation.Account
		for _, simAcc := range accs {
			if simAcc.Address.Equals(provider.Address) {
				simAccount = simAcc
				break
			}
		}
		account := ak.GetAccount(ctx, simAccount.Address)

		msg := types.NewMsgWithdrawRewards(simAccount.Address)

		fees := sdk.Coins{}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		if _, _, err := app.Deliver(tx); err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgWithdrawForeignRewards generates a MsgWithdrawForeignRewards object with all of its fields randomized.
func SimulateMsgWithdrawForeignRewards(k keeper.Keeper, ak types.AccountKeeper, sk types.StakingKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		provider, found := keeper.RandomProvider(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		var simAccount simulation.Account
		for _, simAcc := range accs {
			if simAcc.Address.Equals(provider.Address) {
				simAccount = simAcc
				break
			}
		}
		account := ak.GetAccount(ctx, simAccount.Address)
		toAddr := simulation.RandStringOfLength(r, 42)
		if provider.Rewards.Foreign.Empty() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		denom := provider.Rewards.Foreign[0].Denom
		msg := types.NewMsgWithdrawForeignRewards(provider.Address, denom, toAddr)

		fees := sdk.Coins{}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		if _, _, err := app.Deliver(tx); err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

/*
// SimulateMsgPurchaseShield generates a MsgPurchaseShield object with all of its fields randomized.
func SimulateMsgPurchaseShield(k keeper.Keeper, ak types.AccountKeeper, sk types.StakingKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		purchaser, _ := simulation.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, purchaser.Address)
		bondDenom := sk.BondDenom(ctx)

		poolID, _, found := keeper.RandomPoolInfo(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		pool, found := k.GetPool(ctx, poolID)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		maxPurchaseAmount := sdk.MinInt(pool.Available, account.SpendableCoins(ctx.BlockTime()).AmountOf(bondDenom))
		shieldAmount, err := simulation.RandPositiveInt(r, maxPurchaseAmount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		description := simulation.RandStringOfLength(r, 100)
		claimParams := k.GetClaimProposalParams(ctx)
		shieldEnd := ctx.BlockTime().Add(claimParams.ClaimPeriod).Add(k.GetVotingParams(ctx).VotingPeriod * 2)
		if shieldEnd.After(pool.EndTime) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		if purchaser.Address.Equals(pool.SponsorAddr) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		msg := types.NewMsgPurchaseShield(poolID, sdk.NewCoins(sdk.NewCoin(bondDenom, shieldAmount)), description, purchaser.Address)

		fees := sdk.Coins{}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			purchaser.PrivKey,
		)

		if _, _, err := app.Deliver(tx); err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper, sk types.StakingKeeper) []simulation.WeightedProposalContent {
	return []simulation.WeightedProposalContent{
		{
			AppParamsKey:       OpWeightShieldClaimProposal,
			DefaultWeight:      DefaultWeightShieldClaimProposal,
			ContentSimulatorFn: SimulateShieldClaimProposalContent(k, sk),
		},
	}
}

// SimulateShieldClaimProposalContent generates random shield claim proposal content
func SimulateShieldClaimProposalContent(k keeper.Keeper, sk types.StakingKeeper) simulation.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simulation.Account) govtypes.Content {
		bondDenom := sk.BondDenom(ctx)
		purchaseList, found := keeper.RandomPurchaseList(r, k, ctx)
		if !found {
			return nil
		}
		pool, found := k.GetPool(ctx, purchaseList.PoolID)
		if !found {
			return nil
		}
		if pool.SponsorAddr.Equals(purchaseList.Purchaser) {
			return nil
		}
		entryIndex := r.Intn(len(purchaseList.Entries))
		entry := purchaseList.Entries[entryIndex]
		if entry.ClaimPeriodEndTime.Before(ctx.BlockTime()) {
			return nil
		}
		lossAmount, err := simulation.RandPositiveInt(r, entry.Shield.AmountOf(bondDenom))
		if err != nil {
			return nil
		}

		return types.NewShieldClaimProposal(
			purchaseList.PoolID,
			sdk.NewCoins(sdk.NewCoin(bondDenom, lossAmount)),
			entry.PurchaseID,
			simulation.RandStringOfLength(r, 500),
			simulation.RandStringOfLength(r, 500),
			purchaseList.Purchaser,
		)
	}
}

*/
