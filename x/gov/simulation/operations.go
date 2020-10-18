package simulation

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	"github.com/certikfoundation/shentu/x/gov/internal/keeper"
	"github.com/certikfoundation/shentu/x/gov/internal/types"
	"github.com/certikfoundation/shentu/x/shield"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak govTypes.AccountKeeper, ck types.CertKeeper,
	k keeper.Keeper, wContents []simulation.WeightedProposalContent) simulation.WeightedOperations {

	// generate the weighted operations for the proposal contents
	var wProposalOps simulation.WeightedOperations

	for _, wContent := range wContents {
		wContent := wContent // pin variable
		var weight int
		appParams.GetOrGenerate(cdc, wContent.AppParamsKey, &weight, nil,
			func(_ *rand.Rand) { weight = wContent.DefaultWeight })

		wProposalOps = append(
			wProposalOps,
			simulation.NewWeightedOperation(
				weight,
				SimulateSubmitProposal(ak, ck, k, wContent.ContentSimulatorFn),
			),
		)
	}

	return wProposalOps
}

// SimulateSubmitProposal simulates creating a msg Submit Proposal
// voting on the proposal, and subsequently slashing the proposal. It is implemented using
// future operations.
func SimulateSubmitProposal(
	ak govTypes.AccountKeeper, ck types.CertKeeper, k keeper.Keeper, contentSim simulation.ContentSimulatorFn,
) simulation.Operation {
	// The states are:
	// column 1: All validators vote
	// column 2: 90% vote
	// column 3: 75% vote
	// column 4: 40% vote
	// column 5: 15% vote
	// column 6: noone votes
	// All columns sum to 100 for simplicity, values chosen by @valardragon semi-arbitrarily,
	// feel free to change.
	numVotesTransitionMatrix, _ := simulation.CreateTransitionMatrix([][]int{
		{20, 10, 0, 0, 0, 0},
		{55, 50, 20, 10, 0, 0},
		{25, 25, 30, 25, 30, 15},
		{0, 15, 30, 25, 30, 30},
		{0, 0, 20, 30, 30, 30},
		{0, 0, 0, 10, 10, 25},
	})

	statePercentageArray := []float64{1, .9, .75, .4, .15, 0}
	curNumVotesState := 1

	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		// 1) submit proposal now
		content := contentSim(r, ctx, accs)
		if content == nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
		}

		var (
			deposit sdk.Coins
			err     error
		)
		var simAccount simulation.Account
		if content.ProposalType() == "ShieldClaim" {
			c := content.(shield.ClaimProposal)
			for _, simAcc := range accs {
				if simAcc.Address.Equals(c.Proposer) {
					simAccount = simAcc
					break
				}
			}
			account := ak.GetAccount(ctx, simAccount.Address)
			if account.GetCoins() == nil {
				return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
			}
			denom := account.GetCoins()[0].Denom
			lossAmountDec := c.Loss.AmountOf(denom).ToDec()
			claimProposalParams := k.ShieldKeeper.GetClaimProposalParams(ctx)
			depositRate := claimProposalParams.DepositRate
			minDepositAmountDec := sdk.MaxDec(claimProposalParams.MinDeposit.AmountOf(denom).ToDec(), lossAmountDec.Mul(depositRate))
			minDepositAmount := minDepositAmountDec.Ceil().RoundInt()
			if minDepositAmount.GT(account.SpendableCoins(ctx.BlockTime()).AmountOf(denom)) {
				return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
			}
			deposit = sdk.NewCoins(sdk.NewCoin(denom, minDepositAmount))
		} else {
			simAccount, _ = simulation.RandomAcc(r, accs)
			account := ak.GetAccount(ctx, simAccount.Address)
			spendable := account.SpendableCoins(ctx.BlockTime())
			minDeposit := k.GetDepositParams(ctx).MinDeposit
			if spendable.AmountOf(sdk.DefaultBondDenom).LT(minDeposit.AmountOf(sdk.DefaultBondDenom)) {
				deposit, err = simulation.RandomFees(r, ctx, spendable)
			} else {
				deposit, err = simulation.RandomFees(r, ctx, minDeposit)
			}
			if err != nil {
				return simulation.NoOpMsg(govTypes.ModuleName), nil, err
			}
		}

		minInitialDeposit := k.GetDepositParams(ctx).MinInitialDeposit
		if deposit.AmountOf(sdk.DefaultBondDenom).LT(minInitialDeposit.AmountOf(sdk.DefaultBondDenom)) &&
			!k.IsCouncilMember(ctx, simAccount.Address) {
			return simulation.NewOperationMsgBasic(govTypes.ModuleName,
				"NoOp: insufficient initial deposit amount, skip this tx", "", false, nil), nil, nil
		}

		msg := govTypes.NewMsgSubmitProposal(content, deposit, simAccount.Address)

		account := ak.GetAccount(ctx, simAccount.Address)
		coins := account.SpendableCoins(ctx.BlockTime())

		var fees sdk.Coins
		coins, hasNeg := coins.SafeSub(deposit)
		if !hasNeg {
			fees, err = simulation.RandomFees(r, ctx, coins)
			if err != nil {
				return simulation.NoOpMsg(govTypes.ModuleName), nil, err
			}
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas*5,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		// get the submitted proposal ID
		proposalID, err := k.GetProposalID(ctx)
		if err != nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		opMsg := simulation.NewOperationMsg(msg, true, "")

		proposal, _ := k.GetProposal(ctx, proposalID)
		fmt.Printf(">>>>>>>>>>>>>> proposal id: %d, type: %s, status: %d, init deposit: %s <<<<<<<<<<<<<<<\n",
			proposalID, content.ProposalType(), proposal.Status, proposal.TotalDeposit)

		var fops []simulation.FutureOperation

		// 2) Schedule deposit operations
		if content.ProposalType() != "ShieldClaim" {
			for i := 0; i < 5; i++ {
				fops = append(fops, simulation.FutureOperation{
					BlockHeight: int(ctx.BlockHeight()) + 5,
					Op:          SimulateMsgDeposit(ak, k, proposalID),
				})
			}
		}

		// 3) Schedule operations for certifier voting
		if content.ProposalType() == "ShieldClaim" ||
			content.ProposalType() == "CertifierUpdate" ||
			content.ProposalType() == upgrade.ProposalTypeSoftwareUpgrade {
			for _, acc := range accs {
				if ck.IsCertifier(ctx, acc.Address) {
					fops = append(fops, simulation.FutureOperation{
						BlockHeight: int(ctx.BlockHeight()) + 10,
						Op:          SimulateCertifierMsgVote(ak, k, acc, proposalID),
					})
				}
			}
		}

		// 4) Schedule operations for validator/delegator voting
		// 4.1) first pick a number of people to vote.
		curNumVotesState = numVotesTransitionMatrix.NextState(r, curNumVotesState)
		numVotes := int(math.Ceil(float64(len(accs)) * statePercentageArray[curNumVotesState]))
		// 4.2) select who votes and when
		whoVotes := r.Perm(len(accs))
		whoVotes = whoVotes[:numVotes]

		for i := 0; i < numVotes; i++ {
			fops = append(fops, simulation.FutureOperation{
				BlockHeight: int(ctx.BlockHeight()) + 15,
				Op:          SimulateMsgVote(ak, k, accs[whoVotes[i]], proposalID),
			})
		}

		return opMsg, fops, nil
	}
}

func SimulateMsgVote(ak govTypes.AccountKeeper, k keeper.Keeper,
	simAccount simulation.Account, proposalID uint64) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		fmt.Printf(">>>>>>>>>>>>>> validator vote, id: %d\n", proposalID)

		proposal, ok := k.GetProposal(ctx, proposalID)
		if !ok {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
		}

		fmt.Printf(">>>>>>>>>>>>>> validator vote, status: %d\n", proposal.Status)

		if proposal.Status != types.StatusValidatorVotingPeriod {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
		}

		option := randomVotingOption(r)

		msg := govTypes.NewMsgVote(simAccount.Address, proposalID, option)

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

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
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		fmt.Printf(">>>>>>>>>>>>>> validator vote success, id: %d, option: %s\n", proposalID, option)

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

func SimulateCertifierMsgVote(ak govTypes.AccountKeeper, k keeper.Keeper,
	simAccount simulation.Account, proposalID uint64) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		fmt.Printf(">>>>>>>>>>>>>> certifier vote, id: %d\n", proposalID)

		proposal, ok := k.GetProposal(ctx, proposalID)
		if !ok {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
		}

		fmt.Printf(">>>>>>>>>>>>>> certifier vote, status: %d\n", proposal.Status)

		if proposal.Status != types.StatusCertifierVotingPeriod {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
		}

		var option govTypes.VoteOption

		if simulation.RandIntBetween(r, 0, 100) < 80 {
			option = govTypes.OptionYes
		} else {
			option = govTypes.OptionNo
		}

		msg := govTypes.NewMsgVote(simAccount.Address, proposalID, option)

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

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
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		fmt.Printf(">>>>>>>>>>>>>> certifier vote success, id: %d, option %d\n", proposalID, option)

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

func SimulateMsgDeposit(ak govTypes.AccountKeeper, k keeper.Keeper, proposalID uint64) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		fmt.Printf(">>>>>>>>>>>>>> deposit, id: %d\n", proposalID)

		proposal, ok := k.GetProposal(ctx, proposalID)
		if !ok {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
		}

		fmt.Printf(">>>>>>>>>>>>>> deposit, status: %d, total deposit: %s\n", proposal.Status, proposal.TotalDeposit)

		if proposal.Status != types.StatusDepositPeriod {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
		}

		simAcc, _ := simulation.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simAcc.Address)
		spendable := acc.SpendableCoins(ctx.BlockTime())
		minDeposit := k.GetDepositParams(ctx).MinDeposit
		var deposit sdk.Coins
		var err error
		if spendable.AmountOf(sdk.DefaultBondDenom).LT(minDeposit.AmountOf(sdk.DefaultBondDenom)) {
			deposit, err = simulation.RandomFees(r, ctx, spendable)
		} else {
			deposit, err = simulation.RandomFees(r, ctx, minDeposit)
		}
		if err != nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		msg := govTypes.NewMsgDeposit(simAcc.Address, proposalID, deposit)

		fees, err := simulation.RandomFees(r, ctx, acc.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{acc.GetAccountNumber()},
			[]uint64{acc.GetSequence()},
			simAcc.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		fmt.Printf(">>>>>>>>>>>>>> deposit success, id: %d, amount: %s\n", proposalID, deposit)

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// Pick a random voting option
func randomVotingOption(r *rand.Rand) govTypes.VoteOption {
	switch r.Intn(4) {
	case 0:
		return govTypes.OptionYes
	case 1:
		return govTypes.OptionAbstain
	case 2:
		return govTypes.OptionNo
	case 3:
		return govTypes.OptionNoWithVeto
	default:
		panic("invalid vote option")
	}
}
