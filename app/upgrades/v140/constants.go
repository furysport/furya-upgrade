package v140

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/furysport/furya-upgrade/app/upgrades"
	intertxtypes "github.com/furysport/furya-upgrade/x/intertx/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v1.4.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			icacontrollertypes.StoreKey,
			intertxtypes.StoreKey,
		},
	},
}
