package v6_5_0

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/tfm-com/composable/app/upgrades"
	ibctransfermiddleware "github.com/tfm-com/composable/x/ibctransfermiddleware/types"
)

const (
	// UpgradeName defines the on-chain upgrade name for the composable upgrade.
	UpgradeName = "v6_5_0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{ibctransfermiddleware.StoreKey},
		Deleted: []string{},
	},
}
