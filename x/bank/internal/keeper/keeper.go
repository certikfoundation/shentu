// Package keeper implements custom bank keeper through CVM.
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/hyperledger/burrow/crypto"

	"github.com/certikfoundation/shentu/x/bank/internal/types"
)

// Keeper is a wrapper of the basekeeper with CVM keeper.
type Keeper struct {
	bankKeeper.BaseKeeper
	cvmk types.CVMKeeper
	ak   types.AccountKeeper
}

// NewKeeper returns a new Keeper.
func NewKeeper(
	ak types.AccountKeeper, cvmk types.CVMKeeper, paramSpace params.Subspace, blacklistedAddrs map[string]bool,
) Keeper {
	bk := bankKeeper.NewBaseKeeper(ak, paramSpace, blacklistedAddrs)
	return Keeper{
		BaseKeeper: bk,
		cvmk:       cvmk,
		ak:         ak,
	}
}

// GetCode retrieves VM code from an account.
func (k Keeper) GetCode(ctx sdk.Context, addr sdk.AccAddress) ([]byte, error) {
	vmAddress := crypto.MustAddressFromBytes(addr)
	return k.cvmk.GetCode(ctx, vmAddress)
}

// SendCoins checks if there is code in the receiver account, and wires the send through CVM if it does.
func (k Keeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	code, err := k.GetCode(ctx, toAddr)
	if err != nil {
		return err
	}
	if len(code) > 0 {
		return k.cvmk.Send(ctx, fromAddr, toAddr, amt)
	}
	return k.BaseKeeper.SendCoins(ctx, fromAddr, toAddr, amt)
}
