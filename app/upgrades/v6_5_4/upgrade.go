package v6_5_4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tfm-com/composable/app/keepers"

	"github.com/tfm-com/composable/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	_ codec.Codec,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		BrokenProposals := [3]uint64{2, 6, 11}
		// gov module store
		store := ctx.KVStore(keepers.GetKVStoreKey()["gov"])

		for _, proposalID := range BrokenProposals {
			bz := store.Get(types.ProposalKey(proposalID))
			if bz != nil {
				store.Delete(types.ProposalKey(proposalID))
			}
		}
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
