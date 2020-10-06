package simulation

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/certikfoundation/shentu/common"
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
	DefaultWeightMsgCreatePool             int = 10
	DefaultWeightMsgUpdatePool             int = 10
	DefaultWeightMsgClearPayouts           int = 5
	DefaultWeightMsgDepositCollateral      int = 20
	DefaultWeightMsgWithdrawCollateral     int = 20
	DefaultWeightMsgWithdrawRewards        int = 10
	DefaultWeightMsgWithdrawForeignRewards int = 10
	DefaultWeightMsgPurchaseShield         int = 20
	DefaultWeightShieldClaimProposal       int = 5

	DefaultIntMax            int = 1000000000000
	DefaultTimeOfCoverageMin int = 4838401
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper,
	k keeper.Keeper) simulation.WeightedOperations {
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
	var weightMsgClearPayouts int
	appParams.GetOrGenerate(cdc, OpWeightMsgClearPayouts, &weightMsgClearPayouts, nil,
		func(_ *rand.Rand) {
			weightMsgClearPayouts = DefaultWeightMsgClearPayouts
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
		simulation.NewWeightedOperation(weightMsgCreatePool, SimulateMsgCreatePool(k, ak)),
		simulation.NewWeightedOperation(weightMsgUpdatePool, SimulateMsgUpdatePool(k, ak)),
		simulation.NewWeightedOperation(weightMsgClearPayouts, SimulateMsgClearPayouts(k, ak)),
		simulation.NewWeightedOperation(weightMsgDepositCollateral, SimulateMsgDepositCollateral(k, ak)),
		simulation.NewWeightedOperation(weightMsgWithdrawCollateral, SimulateMsgWithdrawCollateral(k, ak)),
		simulation.NewWeightedOperation(weightMsgWithdrawRewards, SimulateMsgWithdrawRewards(k, ak)),
		simulation.NewWeightedOperation(weightMsgWithdrawForeignRewards, SimulateMsgWithdrawForeignRewards(k, ak)),
		simulation.NewWeightedOperation(weightMsgPurchaseShield, SimulateMsgPurchaseShield(k, ak)),
	}
}

// SimulateMsgCreatePool generates a MsgCreatePool object with all of its fields randomized.
func SimulateMsgCreatePool(k keeper.Keeper, ak types.AccountKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		// admin
		var (
			adminAddr  sdk.AccAddress
			available  sdk.Int
			found      bool
			simAccount simulation.Account
		)
		providers := k.GetAllProviders(ctx)
		if len(providers) == 0 {
			adminAddr, available, found = keeper.RandomDelegation(r, k, ctx)
			if !found {
				return simulation.NoOpMsg(types.ModuleName), nil, nil
			}
			k.SetAdmin(ctx, adminAddr)
		} else {
			adminAddr = k.GetAdmin(ctx)
		}
		for _, simAcc := range accs {
			if simAcc.Address.Equals(adminAddr) {
				simAccount = simAcc
				break
			}
		}
		account := ak.GetAccount(ctx, simAccount.Address)

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
		shield := sdk.NewCoins(sdk.NewCoin(common.MicroCTKDenom, shieldAmount))

		// sponsor
		sponsor := strings.ToLower(simulation.RandStringOfLength(r, 3))

		// deposit
		nativeAmount := account.SpendableCoins(ctx.BlockTime()).AmountOf(common.MicroCTKDenom)
		if !nativeAmount.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		nativeAmount, err = simulation.RandPositiveInt(r, nativeAmount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		nativeDeposit := sdk.NewCoins(sdk.NewCoin(common.MicroCTKDenom, nativeAmount))
		foreignAmount, err := simulation.RandPositiveInt(r, sdk.NewInt(int64(DefaultIntMax)))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		foreignDeposit := sdk.NewCoins(sdk.NewCoin(sponsor, foreignAmount))
		deposit := types.MixedCoins{Native: nativeDeposit, Foreign: foreignDeposit}

		// time of coverage
		timeOfCoverage := int64(simulation.RandIntBetween(r, DefaultTimeOfCoverageMin, DefaultIntMax))

		msg := types.NewMsgCreatePool(simAccount.Address, shield, deposit, sponsor, timeOfCoverage, timeOfCoverage)

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

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgUpdatePool generates a MsgUpdatePool object with all of its fields randomized.
func SimulateMsgUpdatePool(k keeper.Keeper, ak types.AccountKeeper) simulation.Operation {
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
		shield := sdk.NewCoins(sdk.NewCoin(common.MicroCTKDenom, shieldAmount))

		// deposit
		nativeAmount := account.SpendableCoins(ctx.BlockTime()).AmountOf(common.MicroCTKDenom)
		if !nativeAmount.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		nativeAmount, err = simulation.RandPositiveInt(r, nativeAmount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		nativeDeposit := sdk.NewCoins(sdk.NewCoin(common.MicroCTKDenom, nativeAmount))
		foreignAmount, err := simulation.RandPositiveInt(r, sdk.NewInt(int64(DefaultIntMax)))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		foreignDeposit := sdk.NewCoins(sdk.NewCoin(sponsor, foreignAmount))
		deposit := types.MixedCoins{Native: nativeDeposit, Foreign: foreignDeposit}

		// time of coverage
		timeOfCoverage := int64(simulation.RandIntBetween(r, DefaultTimeOfCoverageMin, DefaultIntMax))

		msg := types.NewMsgUpdatePool(simAccount.Address, shield, deposit, poolID, timeOfCoverage, timeOfCoverage)

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

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgClearPayouts generates a MsgClearPayouts object with all of its fields randomized.
func SimulateMsgClearPayouts(k keeper.Keeper, ak types.AccountKeeper) simulation.Operation {
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

		// poolID and sponsor
		_, sponsor, found := keeper.RandomPoolInfo(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		earnings := k.GetPendingPayouts(ctx, sponsor)
		if earnings == nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		msg := types.NewMsgClearPayouts(adminAddr, sponsor)

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

		_, _, err := app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgDepositCollateral generates a MsgDepositCollateral object with all of its fields randomized.
func SimulateMsgDepositCollateral(k keeper.Keeper, ak types.AccountKeeper) simulation.Operation {
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

		// poolID
		poolID, _, found := keeper.RandomPoolInfo(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		// collateral
		provider, found := k.GetProvider(ctx, simAccount.Address)
		if found {
			delAmount = provider.DelegationBonded.AmountOf(common.MicroCTKDenom).Sub(provider.Collateral.AmountOf(common.MicroCTKDenom))
			if !delAmount.IsPositive() {
				return simulation.NoOpMsg(types.ModuleName), nil, nil
			}
		}
		collateralAmount, err := simulation.RandPositiveInt(r, delAmount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		collateral := sdk.NewCoins(sdk.NewCoin(common.MicroCTKDenom, collateralAmount))

		msg := types.NewMsgDepositCollateral(simAccount.Address, poolID, collateral)

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

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgWithdrawCollateral generates a MsgWithdrawCollateral object with all of its fields randomized.
func SimulateMsgWithdrawCollateral(k keeper.Keeper, ak types.AccountKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		collateral, found := keeper.RandomCollateral(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		var simAccount simulation.Account
		for _, simAcc := range accs {
			if simAcc.Address.Equals(collateral.Provider) {
				simAccount = simAcc
				break
			}
		}
		account := ak.GetAccount(ctx, simAccount.Address)

		collateralNotBeingWithdrawn := collateral.Amount.Sub(collateral.Withdrawal)
		withdrawalAmount, err := simulation.RandPositiveInt(r, collateralNotBeingWithdrawn.AmountOf(common.MicroCTKDenom))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		withdrawal := sdk.NewCoins(sdk.NewCoin(common.MicroCTKDenom, withdrawalAmount))

		msg := types.NewMsgWithdrawCollateral(simAccount.Address, collateral.PoolID, withdrawal)

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

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgWithdrawRewards generates a MsgWithdrawRewards object with all of its fields randomized.
func SimulateMsgWithdrawRewards(k keeper.Keeper, ak types.AccountKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		collateral, found := keeper.RandomCollateral(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		var simAccount simulation.Account
		for _, simAcc := range accs {
			if simAcc.Address.Equals(collateral.Provider) {
				simAccount = simAcc
				break
			}
		}
		account := ak.GetAccount(ctx, collateral.Provider)

		msg := types.NewMsgWithdrawRewards(collateral.Provider)

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

		_, _, err := app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgWithdrawForeignRewards generates a MsgWithdrawForeignRewards object with all of its fields randomized.
func SimulateMsgWithdrawForeignRewards(k keeper.Keeper, ak types.AccountKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		collateral, found := keeper.RandomCollateral(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		var simAccount simulation.Account
		for _, simAcc := range accs {
			if simAcc.Address.Equals(collateral.Provider) {
				simAccount = simAcc
				break
			}
		}
		account := ak.GetAccount(ctx, collateral.Provider)

		pool, err := k.GetPool(ctx, collateral.PoolID)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		toAddr := simulation.RandStringOfLength(r, 42)

		if !k.GetRewards(ctx, collateral.Provider).Foreign.AmountOf(pool.Sponsor).IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		msg := types.NewMsgWithdrawForeignRewards(collateral.Provider, pool.Sponsor, toAddr)

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

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgPurchaseShield generates a MsgPurchaseShield object with all of its fields randomized.
func SimulateMsgPurchaseShield(k keeper.Keeper, ak types.AccountKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		purchaser, _ := simulation.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, purchaser.Address)

		poolID, _, found := keeper.RandomPoolInfo(r, k, ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		pool, err := k.GetPool(ctx, poolID)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		maxPurchaseAmount := sdk.MinInt(pool.Available, account.SpendableCoins(ctx.BlockTime()).AmountOf(common.MicroCTKDenom))
		shieldAmount, err := simulation.RandPositiveInt(r, maxPurchaseAmount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		description := simulation.RandStringOfLength(r, 100)

		msg := types.NewMsgPurchaseShield(poolID, sdk.NewCoins(sdk.NewCoin(common.MicroCTKDenom, shieldAmount)), description, purchaser.Address)

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

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []simulation.WeightedProposalContent {
	return []simulation.WeightedProposalContent{
		{
			AppParamsKey:       OpWeightShieldClaimProposal,
			DefaultWeight:      DefaultWeightShieldClaimProposal,
			ContentSimulatorFn: SimulateShieldClaimProposalContent(k),
		},
	}
}

// SimulateShieldClaimProposalContent generates random shield claim proposal content
func SimulateShieldClaimProposalContent(k keeper.Keeper) simulation.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simulation.Account) govtypes.Content {
		// fmt.Printf(">> debug SimulateShieldClaimProposalContent <<\n")
		purchase, found := keeper.RandomPurchase(r, k, ctx)
		if !found {
			// fmt.Printf(">> debug SimulateShieldClaimProposalContent: purchase is not found <<\n")
			return nil
		}
		lossAmount, err := simulation.RandPositiveInt(r, purchase.Shield.AmountOf(common.MicroCTKDenom))
		if err != nil {
			fmt.Printf(">> debug SimulateShieldClaimProposalContent: invalid shield %s <<\n", purchase.Shield)
			return nil
		}
		txhash := hex.EncodeToString(purchase.TxHash)

		fmt.Printf(">> debug SimulateShieldClaimProposalContent: pool id %d, loss %s, proposer %s\n",
			purchase.PoolID, lossAmount, purchase.Purchaser)

		return types.NewShieldClaimProposal(
			purchase.PoolID,
			sdk.NewCoins(sdk.NewCoin(common.MicroCTKDenom, lossAmount)),
			simulation.RandStringOfLength(r, 500),
			txhash,
			simulation.RandStringOfLength(r, 500),
			purchase.Purchaser,
		)
	}
}
