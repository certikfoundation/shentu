package simulation

import (
	"math"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	"github.com/certikfoundation/shentu/x/cert"
	"github.com/certikfoundation/shentu/x/gov/internal/keeper"
	"github.com/certikfoundation/shentu/x/gov/internal/types"
	"github.com/certikfoundation/shentu/x/shield"
)

var initialProposalID = uint64(100000000000000)

// Simulation operation weights constants
const (
	OpWeightMsgDeposit = "op_weight_msg_deposit"
	OpWeightMsgVote    = "op_weight_msg_vote"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak govTypes.AccountKeeper, ck types.CertKeeper,
	k keeper.Keeper, wContents []simulation.WeightedProposalContent) simulation.WeightedOperations {
	var (
		weightMsgDeposit int
		weightMsgVote    int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgDeposit, &weightMsgDeposit, nil,
		func(_ *rand.Rand) {
			weightMsgDeposit = simappparams.DefaultWeightMsgDeposit
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgVote, &weightMsgVote, nil,
		func(_ *rand.Rand) {
			weightMsgVote = simappparams.DefaultWeightMsgVote
		},
	)

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

	wGovOps := simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgDeposit,
			SimulateMsgDeposit(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgVote,
			SimulateMsgVote(ak, k),
		),
	}

	return append(wProposalOps, wGovOps...)
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
			skip    bool
			err     error
		)
		var simAccount simulation.Account
		if content.ProposalType() == shield.ProposalTypeShieldClaim {
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
			deposit, skip, err = randomDeposit(r, ctx, ak, k, simAccount.Address)
			switch {
			case skip:
				return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
			case err != nil:
				return simulation.NoOpMsg(govTypes.ModuleName), nil, err
			}
		}

		minInitialDeposit := k.GetDepositParams(ctx).MinInitialDeposit
		if deposit.AmountOf(sdk.DefaultBondDenom).LT(minInitialDeposit.AmountOf(sdk.DefaultBondDenom)) &&
			!k.IsCouncilMember(ctx, simAccount.Address) {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, nil
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

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		opMsg := simulation.NewOperationMsg(msg, true, "")

		// get the submitted proposal ID
		proposalID, err := k.GetProposalID(ctx)
		if err != nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		// 2) Schedule operations for votes
		// 2.1) first pick a number of people to vote.
		curNumVotesState = numVotesTransitionMatrix.NextState(r, curNumVotesState)
		numVotes := int(math.Ceil(float64(len(accs)) * statePercentageArray[curNumVotesState]))

		// 2.2) select who votes and when
		whoVotes := r.Perm(len(accs))

		// didntVote := whoVotes[numVotes:]
		whoVotes = whoVotes[:numVotes]
		votingPeriod := k.GetVotingParams(ctx).VotingPeriod

		var fops []simulation.FutureOperation

		if content.ProposalType() == shield.ProposalTypeShieldClaim ||
			content.ProposalType() == cert.ProposalTypeCertifierUpdate ||
			content.ProposalType() == upgrade.ProposalTypeSoftwareUpgrade {
			// certifier voting
			for _, acc := range accs {
				if ck.IsCertifier(ctx, acc.Address) {
					whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second / 4)
					fops = append(fops, simulation.FutureOperation{
						BlockTime: whenVote,
						Op:        certifierSimulateMsgVote(ak, acc, proposalID),
					})
				}
			}
		}

		// validator / delegator voting
		for i := 0; i < numVotes; i++ {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops = append(fops, simulation.FutureOperation{
				BlockTime: whenVote,
				Op:        operationSimulateMsgVote(ak, k, accs[whoVotes[i]], int64(proposalID)),
			})
		}

		return opMsg, fops, nil
	}
}

// SimulateMsgDeposit generates a MsgDeposit with random values.
func SimulateMsgDeposit(ak govTypes.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		simAccount, _ := simulation.RandomAcc(r, accs)
		proposalID, ok := randomProposalID(r, k, ctx, types.StatusDepositPeriod)
		if !ok {
			return simulation.NewOperationMsgBasic(govTypes.ModuleName,
				"NoOp: randomly selected proposal not in deposit period, skip this tx", "", false, nil), nil, nil
		}

		deposit, skip, err := randomDeposit(r, ctx, ak, k, simAccount.Address)
		switch {
		case skip:
			return simulation.NewOperationMsgBasic(govTypes.ModuleName,
				"NoOp: zero balance, skip this tx", "", false, nil), nil, nil
		case err != nil:
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		msg := govTypes.NewMsgDeposit(simAccount.Address, proposalID, deposit)

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

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgVote generates a MsgVote with random values.
func SimulateMsgVote(ak govTypes.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return operationSimulateMsgVote(ak, k, simulation.Account{}, -1)
}

func operationSimulateMsgVote(ak govTypes.AccountKeeper, k keeper.Keeper,
	simAccount simulation.Account, proposalIDInt int64) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		if simAccount.Equals(simulation.Account{}) {
			simAccount, _ = simulation.RandomAcc(r, accs)
		}

		var proposalID uint64

		switch {
		case proposalIDInt < 0:
			var ok bool
			proposalID, ok = randomProposalID(r, k, ctx, types.StatusValidatorVotingPeriod)
			if !ok {
				return simulation.NewOperationMsgBasic(govTypes.ModuleName,
					"NoOp: randomly selected proposal not in validator voting period, skip this tx", "", false, nil), nil, nil
			}
		default:
			proposalID = uint64(proposalIDInt)
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

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

func certifierSimulateMsgVote(ak govTypes.AccountKeeper, certifier simulation.Account, proposalID uint64) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		var option govTypes.VoteOption

		if simulation.RandIntBetween(r, 0, 100) < 75 {
			option = govTypes.OptionYes
		} else {
			option = govTypes.OptionNo
		}

		msg := govTypes.NewMsgVote(certifier.Address, proposalID, option)

		account := ak.GetAccount(ctx, certifier.Address)
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
			certifier.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(govTypes.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// Pick a random deposit with a random denomination with a
// deposit amount between (0, min(balance, minDepositAmount))
// This is to simulate multiple users depositing to get the
// proposal above the minimum deposit amount
func randomDeposit(r *rand.Rand, ctx sdk.Context,
	ak govTypes.AccountKeeper, k keeper.Keeper, addr sdk.AccAddress,
) (deposit sdk.Coins, skip bool, err error) {
	account := ak.GetAccount(ctx, addr)
	coins := account.SpendableCoins(ctx.BlockHeader().Time)
	if coins.Empty() {
		return nil, true, nil // skip
	}

	minDeposit := k.GetDepositParams(ctx).MinDeposit
	denomIndex := r.Intn(len(minDeposit))
	denom := minDeposit[denomIndex].Denom

	depositCoins := coins.AmountOf(denom)
	if depositCoins.IsZero() {
		return nil, true, nil
	}

	maxAmt := depositCoins
	if maxAmt.GT(minDeposit[denomIndex].Amount) {
		maxAmt = minDeposit[denomIndex].Amount
	}

	amount, err := simulation.RandPositiveInt(r, maxAmt)
	if err != nil {
		return nil, false, err
	}

	return sdk.Coins{sdk.NewCoin(denom, amount)}, false, nil
}

// Pick a random proposal ID between the initial proposal ID
// (defined in gov GenesisState) and the latest proposal ID
// that matches a given Status.
// It does not provide a default ID.
func randomProposalID(r *rand.Rand, k keeper.Keeper,
	ctx sdk.Context, status types.ProposalStatus) (proposalID uint64, found bool) {
	proposalID, _ = k.GetProposalID(ctx)

	switch {
	case proposalID > initialProposalID:
		// select a random ID between [initialProposalID, proposalID]
		proposalID = uint64(simulation.RandIntBetween(r, int(initialProposalID), int(proposalID)))

	default:
		// This is called on the first call to this funcion
		// in order to update the global variable
		initialProposalID = proposalID
	}

	proposal, ok := k.GetProposal(ctx, proposalID)
	if !ok || proposal.Status != status {
		return proposalID, false
	}

	return proposalID, true
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
