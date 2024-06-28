package keeper

import (
	"github.com/cometbft/cometbft/libs/log"
	"github.com/tfm-com/composable/x/ibctransfermiddleware/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the staking middleware store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	addresses []string
}

// NewKeeper creates a new middleware Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	authority string,
	addresses []string,
) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  key,
		authority: authority,
		addresses: addresses,
	}
}

// GetAuthority returns the x/ibctransfermiddleware module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetParams sets the x/ibctransfermiddleware module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)
	return nil
}

// GetParams returns the current x/ibctransfermiddleware module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

func (k Keeper) GetSequenceFee(ctx sdk.Context, sequence uint64) (coin sdk.Coin, found bool) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetSequenceKey(sequence))
	if value == nil {
		return sdk.Coin{}, false
	}

	fee := types.MustUnmarshalCoin(k.cdc, value)
	return fee, true
}

func (k Keeper) SetSequenceFee(ctx sdk.Context, sequence uint64, coin sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetSequenceKey(sequence), types.MustMarshalCoin(k.cdc, &coin))
}

func (k Keeper) DeleteSequenceFee(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetSequenceKey(sequence))
}

func (k Keeper) GetCoin(ctx sdk.Context, targetChannelID, denom string) *types.CoinItem {
	params := k.GetParams(ctx)
	channelFee := findChannelParams(params.ChannelFees, targetChannelID)
	if channelFee == nil {
		return nil
	}
	return findCoinByDenom(channelFee.AllowedTokens, denom)
}

func (k Keeper) GetChannelFeeAddress(ctx sdk.Context, targetChannelID string) string {
	params := k.GetParams(ctx)
	channelFee := findChannelParams(params.ChannelFees, targetChannelID)
	if channelFee == nil {
		return ""
	}
	return channelFee.FeeAddress
}
