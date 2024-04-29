package v6_6_2_test

import (
	"encoding/json"
	"testing"

	"github.com/notional-labs/composable/v6/app/upgrades/v6_6_2"
	ibchookskeeper "github.com/notional-labs/composable/v6/x/ibc-hooks/keeper"
	ibctransfermiddlewaretypes "github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	apptesting "github.com/notional-labs/composable/v6/app"
	"github.com/notional-labs/composable/v6/bech32-migration/utils"
	ibchookstypes "github.com/notional-labs/composable/v6/x/ibc-hooks/types"
	"github.com/stretchr/testify/suite"
)

const (
	COIN_DENOM   = "stake"
	CONNECTION_0 = "connection-0"
	PORT_0       = "port-0"
	CHANNEL_0    = "channel-0"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// Ensures the test does not error out.
func (s *UpgradeTestSuite) TestForMigratingNewPrefix() {
	// DEFAULT PREFIX: centauri
	sdk.SetAddrCacheEnabled(false)

	sdk.GetConfig().SetBech32PrefixForAccount(utils.OldBech32PrefixAccAddr, utils.OldBech32PrefixAccPub)
	sdk.GetConfig().SetBech32PrefixForValidator(utils.OldBech32PrefixValAddr, utils.OldBech32PrefixValPub)
	sdk.GetConfig().SetBech32PrefixForConsensusNode(utils.OldBech32PrefixConsAddr, utils.OldBech32PrefixConsPub)

	s.Setup(s.T())

	prepareForTestingIbcTransferMiddlewareModule(s)
	prepareForTestingIbcHooksModule(s)

	/* == UPGRADE == */
	upgradeHeight := int64(5)
	s.ConfirmUpgradeSucceeded(v6_6_2.UpgradeName, upgradeHeight)

	checkUpgradeIbcTransferMiddlewareModule(s)
	checkUpgradeIbcHooksMiddlewareModule(s)
}

func prepareForTestingIbcTransferMiddlewareModule(s *UpgradeTestSuite) {
	store := s.Ctx.KVStore(s.App.GetKey(ibctransfermiddlewaretypes.StoreKey))
	var fees []*ibctransfermiddlewaretypes.ChannelFee
	fees = append(fees, &ibctransfermiddlewaretypes.ChannelFee{
		Channel: "channel-7",
		AllowedTokens: []*ibctransfermiddlewaretypes.CoinItem{{
			MinFee:     sdk.Coin{},
			Percentage: 20,
		}},
		FeeAddress:          "centauri1hj5fveer5cjtn4wd6wstzugjfdxzl0xpzxlwgs",
		MinTimeoutTimestamp: 0,
	})
	fees = append(fees, &ibctransfermiddlewaretypes.ChannelFee{
		Channel: "channel-9",
		AllowedTokens: []*ibctransfermiddlewaretypes.CoinItem{{
			MinFee:     sdk.Coin{},
			Percentage: 10,
		}},
		FeeAddress:          "centauri1hj5fveer5cjtn4wd6wstzugjfdxzl0xpzxlwgs",
		MinTimeoutTimestamp: 0,
	})
	params := ibctransfermiddlewaretypes.Params{ChannelFees: fees}
	encCdc := apptesting.MakeEncodingConfig()
	bz := encCdc.Amino.MustMarshal(&params)
	store.Set(ibctransfermiddlewaretypes.ParamsKey, bz)
}

func prepareForTestingIbcHooksModule(s *UpgradeTestSuite) {
	store := s.Ctx.KVStore(s.App.GetKey(ibchookstypes.StoreKey))
	store.Set(ibchookskeeper.GetPacketKey("channel-2", 2), []byte("centauri1hj5fveer5cjtn4wd6wstzugjfdxzl0xpzxlwgs"))
	store.Set(ibchookskeeper.GetPacketKey("channel-4", 2), []byte("centauri1wkjvpgkuchq0r8425g4z4sf6n85zj5wtmqzjv9"))
}

func checkUpgradeIbcTransferMiddlewareModule(s *UpgradeTestSuite) {
	data := s.App.IbcTransferMiddlewareKeeper.GetChannelFeeAddress(s.Ctx, "channel-9")
	s.Suite.Equal("pica1hj5fveer5cjtn4wd6wstzugjfdxzl0xpas3hgy", data)

	data = s.App.IbcTransferMiddlewareKeeper.GetChannelFeeAddress(s.Ctx, "channel-7")
	s.Suite.Equal("pica1hj5fveer5cjtn4wd6wstzugjfdxzl0xpas3hgy", data)
	data = s.App.IbcTransferMiddlewareKeeper.GetChannelFeeAddress(s.Ctx, "channel-1")
	s.Suite.Equal("", data)
}

func checkUpgradeIbcHooksMiddlewareModule(s *UpgradeTestSuite) {
	data := s.App.IBCHooksKeeper.GetPacketCallback(s.Ctx, "channel-2", 2)
	s.Suite.Equal("pica1hj5fveer5cjtn4wd6wstzugjfdxzl0xpas3hgy", data)

	data = s.App.IBCHooksKeeper.GetPacketCallback(s.Ctx, "channel-4", 2)
	s.Suite.Equal("pica1wkjvpgkuchq0r8425g4z4sf6n85zj5wtykvtv3", data)

	data = s.App.IBCHooksKeeper.GetPacketCallback(s.Ctx, "channel-2", 1)
	s.Suite.Equal("", data)
}

func CreateVestingAccount(s *UpgradeTestSuite,
) vestingtypes.ContinuousVestingAccount {
	str := `{"@type":"/cosmos.vesting.v1beta1.ContinuousVestingAccount","base_vesting_account":{"base_account":{"address":"centauri1alga5e8vr6ccr9yrg0kgxevpt5xgmgrvfkc5p8","pub_key":{"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey","threshold":4,"public_keys":[{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AlnzK22KrkylnvTCvZZc8eZnydtQuzCWLjJJSMFUvVHf"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Aiw2Ftg+fnoHDU7M3b0VMRsI0qurXlerW0ahtfzSDZA4"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AvEHv+MVYRVau8FbBcJyG0ql85Tbbn7yhSA0VGmAY4ku"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Az5VHWqi3zMJu1rLGcu2EgNXLLN+al4Dy/lj6UZTzTCl"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Ai4GlSH3uG+joMnAFbQC3jQeHl9FPvVTlRmwIFt7d7TI"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A2kAzH2bZr530jmFq/bRFrT2q8SRqdnfIebba+YIBqI1"}]},"account_number":46,"sequence":27},"original_vesting":[{"denom":"stake","amount":"22165200000000"}],"delegated_free":[{"denom":"stake","amount":"443382497453"}],"delegated_vesting":[{"denom":"stake","amount":"22129422502547"}],"end_time":1770994800},"start_time":1676300400}`

	var acc vestingtypes.ContinuousVestingAccount
	if err := json.Unmarshal([]byte(str), &acc); err != nil {
		panic(err)
	}

	err := banktestutil.FundAccount(s.App.BankKeeper, s.Ctx, acc.BaseAccount.GetAddress(),
		acc.GetOriginalVesting())
	if err != nil {
		panic(err)
	}

	err = banktestutil.FundAccount(s.App.BankKeeper, s.Ctx, acc.BaseAccount.GetAddress(),
		sdk.NewCoins(sdk.NewCoin(COIN_DENOM, math.NewIntFromUint64(1))))
	if err != nil {
		panic(err)
	}

	s.App.AccountKeeper.SetAccount(s.Ctx, &acc)
	return acc
}
