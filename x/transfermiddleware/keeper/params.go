package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xTFM/composable-cosmos/x/transfermiddleware/types"
)

func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
