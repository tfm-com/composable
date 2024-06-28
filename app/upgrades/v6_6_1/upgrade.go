package v6_6_1

import (
	ibchookstypes "github.com/0xTFM/composable-cosmos/x/ibc-hooks/types"
	ibctransfermiddlewaretypes "github.com/0xTFM/composable-cosmos/x/ibctransfermiddleware/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	routertypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"

	"github.com/0xTFM/composable-cosmos/app/keepers"
	"github.com/0xTFM/composable-cosmos/app/upgrades"
	bech32authmigration "github.com/0xTFM/composable-cosmos/bech32-migration/auth"
	bech32govmigration "github.com/0xTFM/composable-cosmos/bech32-migration/gov"
	bech32IbcHooksMigration "github.com/0xTFM/composable-cosmos/bech32-migration/ibchooks"
	bench32ibctransfermiddleware "github.com/0xTFM/composable-cosmos/bech32-migration/ibctransfermiddleware"
	bech32icamigration "github.com/0xTFM/composable-cosmos/bech32-migration/ica"
	bech32mintmigration "github.com/0xTFM/composable-cosmos/bech32-migration/mint"
	bech32PfmMigration "github.com/0xTFM/composable-cosmos/bech32-migration/pfmmiddleware"
	bech32slashingmigration "github.com/0xTFM/composable-cosmos/bech32-migration/slashing"
	bech32stakingmigration "github.com/0xTFM/composable-cosmos/bech32-migration/staking"
	bech32transfermiddlewaremigration "github.com/0xTFM/composable-cosmos/bech32-migration/transfermiddleware"
	bech32WasmMigration "github.com/0xTFM/composable-cosmos/bech32-migration/wasm"
	transfermiddlewaretypes "github.com/0xTFM/composable-cosmos/x/transfermiddleware/types"
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
		bech32stakingmigration.MigrateAddressBech32(ctx, keys[stakingtypes.StoreKey], codec)
		bech32stakingmigration.MigrateUnbonding(ctx, keys[stakingtypes.StoreKey], codec)
		bech32slashingmigration.MigrateAddressBech32(ctx, keys[slashingtypes.StoreKey], codec)
		bech32govmigration.MigrateAddressBech32(ctx, keys[govtypes.StoreKey], codec)
		bech32authmigration.MigrateAddressBech32(ctx, keys[authtypes.StoreKey], codec)
		bech32icamigration.MigrateAddressBech32(ctx, keys[icahosttypes.StoreKey], codec)
		bech32mintmigration.MigrateAddressBech32(ctx, keys[minttypes.StoreKey], codec)
		bech32transfermiddlewaremigration.MigrateAddressBech32(ctx, keys[transfermiddlewaretypes.StoreKey], codec)
		bech32WasmMigration.MigrateAddressBech32(ctx, keys[wasm.StoreKey], codec)
		bech32PfmMigration.MigrateAddressBech32(ctx, keys[routertypes.StoreKey], codec, keepers)
		bench32ibctransfermiddleware.MigrateAddressBech32(ctx, keys[ibctransfermiddlewaretypes.StoreKey], codec)
		bech32IbcHooksMigration.MigrateAddressBech32(ctx, keys[ibchookstypes.StoreKey], codec)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
