package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"

	custombankkeeper "github.com/0xTFM/composable-cosmos/custom/bank/keeper"
	ibctransfermiddleware "github.com/0xTFM/composable-cosmos/x/ibctransfermiddleware/keeper"
	ibctransfermiddlewaretypes "github.com/0xTFM/composable-cosmos/x/ibctransfermiddleware/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

type Keeper struct {
	ibctransferkeeper.Keeper
	cdc                   codec.BinaryCodec
	IbcTransfermiddleware *ibctransfermiddleware.Keeper
	bank                  *custombankkeeper.Keeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper,
	bk types.BankKeeper,
	scopedKeeper exported.ScopedKeeper,
	ibcTransfermiddleware *ibctransfermiddleware.Keeper,
	bankKeeper *custombankkeeper.Keeper,
) Keeper {
	keeper := Keeper{
		Keeper:                ibctransferkeeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, portKeeper, authKeeper, bk, scopedKeeper),
		IbcTransfermiddleware: ibcTransfermiddleware,
		cdc:                   cdc,
		bank:                  bankKeeper,
	}
	return keeper
}

// Transfer is the server API around the Transfer method of the IBC transfer module.
// It checks if the sender is allowed to transfer the token and if the channel has fees.
// If the channel has fees, it will charge the sender and send the fees to the fee address.
// If the sender is not allowed to transfer the token because this tokens does not exists in the allowed tokens list, it just return without doing anything.
// If the sender is allowed to transfer the token, it will call the original transfer method.
// If the transfer amount is less than the minimum fee, it will charge the full transfer amount.
// If the transfer amount is greater than the minimum fee, it will charge the minimum fee and the percentage fee.
func (k Keeper) Transfer(goCtx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.IbcTransfermiddleware.GetParams(ctx)
	charge_coin := sdk.NewCoin(msg.Token.Denom, sdk.ZeroInt())
	if params.ChannelFees != nil && len(params.ChannelFees) > 0 {
		channelFee := findChannelParams(params.ChannelFees, msg.SourceChannel)
		if channelFee != nil {
			if channelFee.MinTimeoutTimestamp > 0 {

				goCtx := sdk.UnwrapSDKContext(goCtx)
				blockTime := goCtx.BlockTime()

				timeoutTimeInFuture := time.Unix(0, int64(msg.TimeoutTimestamp))
				if timeoutTimeInFuture.Before(blockTime) {
					return nil, fmt.Errorf("incorrect timeout timestamp found during ibc transfer. timeout timestamp is in the past")
				}

				difference := timeoutTimeInFuture.Sub(blockTime).Nanoseconds()
				if difference < channelFee.MinTimeoutTimestamp {
					return nil, fmt.Errorf("incorrect timeout timestamp found during ibc transfer. too soon")
				}
			}
			coin := findCoinByDenom(channelFee.AllowedTokens, msg.Token.Denom)
			if coin == nil {
				return nil, fmt.Errorf("token not allowed to be transferred in this channel")
			}

			minFee := coin.MinFee.Amount
			priority := GetPriority(msg.Memo)
			if priority != nil {
				p := findPriority(coin.TxPriorityFee, *priority)
				if p != nil && coin.MinFee.Denom == p.PriorityFee.Denom {
					minFee = minFee.Add(p.PriorityFee.Amount)
				}
			}

			charge := minFee
			if charge.GT(msg.Token.Amount) {
				charge = msg.Token.Amount
			}

			newAmount := msg.Token.Amount.Sub(charge)

			if newAmount.IsPositive() {
				percentageCharge := newAmount.QuoRaw(coin.Percentage)
				newAmount = newAmount.Sub(percentageCharge)
				charge = charge.Add(percentageCharge)
			}

			msgSender, err := sdk.AccAddressFromBech32(msg.Sender)
			if err != nil {
				return nil, err
			}

			feeAddress, err := sdk.AccAddressFromBech32(channelFee.FeeAddress)
			if err != nil {
				return nil, err
			}

			charge_coin = sdk.NewCoin(msg.Token.Denom, charge)
			send_err := k.bank.SendCoins(ctx, msgSender, feeAddress, sdk.NewCoins(charge_coin))
			if send_err != nil {
				return nil, send_err
			}

			if newAmount.LTE(sdk.ZeroInt()) {
				return &types.MsgTransferResponse{}, nil
			}
			msg.Token.Amount = newAmount
		}
	}
	ret, err := k.Keeper.Transfer(goCtx, msg)
	if err == nil && ret != nil && !charge_coin.IsZero() {
		k.IbcTransfermiddleware.SetSequenceFee(ctx, ret.Sequence, charge_coin)
	}
	return ret, err
}

func GetPriority(jsonString string) *string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		return nil
	}

	priority, ok := data["priority"].(string)
	if !ok {
		return nil
	}

	return &priority
}

func findPriority(priorities []*ibctransfermiddlewaretypes.TxPriorityFee, priority string) *ibctransfermiddlewaretypes.TxPriorityFee {
	for _, p := range priorities {
		if p.Priority == priority {
			return p
		}
	}
	return nil
}
