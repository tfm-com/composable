package v6_6_2

import (
	ibchookstypes "github.com/0xTFM/composable-cosmos/x/ibc-hooks/types"
	ibctransfermiddlewaretypes "github.com/0xTFM/composable-cosmos/x/ibctransfermiddleware/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/0xTFM/composable-cosmos/app/keepers"
	"github.com/0xTFM/composable-cosmos/app/upgrades"
	bech32IbcHooksMigration "github.com/0xTFM/composable-cosmos/bech32-migration/ibchooks"
	bench32ibctransfermiddleware "github.com/0xTFM/composable-cosmos/bech32-migration/ibctransfermiddleware"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	codec codec.Codec,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		keys := keepers.GetKVStoreKey()
		// Migration prefix
		ctx.Logger().Info("First step: Migrate addresses stored in bech32 form to use new prefix")
		bench32ibctransfermiddleware.MigrateAddressBech32(ctx, keys[ibctransfermiddlewaretypes.StoreKey], codec)
		bech32IbcHooksMigration.MigrateAddressBech32(ctx, keys[ibchookstypes.StoreKey], codec)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
