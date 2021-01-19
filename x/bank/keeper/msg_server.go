package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	vesting "github.com/certikfoundation/shentu/x/auth/types"
	"github.com/certikfoundation/shentu/x/bank/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) Send(goCtx context.Context, msg *bankTypes.MsgSend) (*bankTypes.MsgSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.SendEnabledCoins(ctx, msg.Amount...); err != nil {
		return nil, err
	}

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if k.BlockedAddr(to) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	}

	err = k.SendCoins(ctx, from, to, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, bankTypes.AttributeValueCategory),
		),
	)

	return &bankTypes.MsgSendResponse{}, nil
}

func (k msgServer) MultiSend(goCtx context.Context, msg *bankTypes.MsgMultiSend) (*bankTypes.MsgMultiSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// NOTE: totalIn == totalOut should already have been checked
	for _, in := range msg.Inputs {
		if err := k.SendEnabledCoins(ctx, in.Coins...); err != nil {
			return nil, err
		}
	}

	for _, out := range msg.Outputs {
		accAddr, err := sdk.AccAddressFromBech32(out.Address)
		if err != nil {
			panic(err)
		}
		if k.BlockedAddr(accAddr) {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive transactions", out.Address)
		}
	}

	err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, bankTypes.AttributeValueCategory),
		),
	)

	return &bankTypes.MsgMultiSendResponse{}, nil
}

func (k msgServer) LockedSend(goCtx context.Context, msg *types.MsgLockedSend) (*types.MsgLockedSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	fromAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}
	unlocker, err := sdk.AccAddressFromBech32(msg.UnlockerAddress)
	if err != nil {
		return nil, err
	}

	// preliminary checks
	from := k.ak.GetAccount(ctx, fromAddr)
	if from == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "sender account %s does not exist", msg.FromAddress)
	}
	if toAddr.Equals(unlocker) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "recipient cannot be the unlocker")
	}

	acc := k.ak.GetAccount(ctx, toAddr)

	var toAcc *vesting.ManualVestingAccount
	if acc == nil {
		acc = k.ak.NewAccountWithAddress(ctx, toAddr)
		toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
		if err != nil {
			panic(err)
		}

		baseAcc := authtypes.NewBaseAccount(toAddr, acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
		if unlocker.Empty() {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid unlocker address provided")
		}
		toAcc = vesting.NewManualVestingAccount(baseAcc, sdk.NewCoins(), sdk.NewCoins(), unlocker)
	} else {
		var ok bool
		toAcc, ok = acc.(*vesting.ManualVestingAccount)
		if !ok {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "receiver account is not a ManualVestingAccount")
		}
		if !unlocker.Empty() {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "cannot change the unlocker for existing ManualVestingAccount")
		}
	}

	// add to receiver account as normally done
	// but make the added amount vesting (OV := Vesting + Vested)
	err = k.AddCoins(ctx, toAddr, msg.Amount)
	if err != nil {
		return nil, err
	}

	toAcc.OriginalVesting = toAcc.OriginalVesting.Add(msg.Amount...)
	k.ak.SetAccount(ctx, toAcc)

	// subtract from sender account (as normally done)
	err = k.SubtractCoins(ctx, fromAddr, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLockedSendToVestingAccount,
			sdk.NewAttribute(bankTypes.AttributeKeySender, msg.FromAddress),
			sdk.NewAttribute(bankTypes.AttributeKeyRecipient, msg.ToAddress),
			sdk.NewAttribute(types.AttributeKeyUnlocker, msg.UnlockerAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
	)
	return &types.MsgLockedSendResponse{}, nil
}
