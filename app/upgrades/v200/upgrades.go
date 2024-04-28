package v200

import (
	"reflect"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/furysport/furya-chain/app/keepers"
	minttypes "github.com/furysport/furya-chain/x/mint/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	alliancemoduletypes "github.com/terra-money/alliance/x/alliance/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

//nolint:all
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("start to run module migrations...")

		for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
			subspace := subspace

			var keyTable paramstypes.KeyTable
			switch subspace.Name() {
			case authtypes.ModuleName:
				keyTable = authtypes.ParamKeyTable()
			case banktypes.ModuleName:
				keyTable = banktypes.ParamKeyTable()
			case stakingtypes.ModuleName:
				keyTable = stakingtypes.ParamKeyTable()
			case slashingtypes.ModuleName:
				keyTable = slashingtypes.ParamKeyTable()
			case crisistypes.ModuleName:
				keyTable = crisistypes.ParamKeyTable()
			case govtypes.ModuleName:
				keyTable = govv1.ParamKeyTable()
			case distrtypes.ModuleName:
				keyTable = distrtypes.ParamKeyTable()
			case wasmtypes.ModuleName:
				keyTable = wasmtypes.ParamKeyTable()
			case minttypes.ModuleName:
				keyTable = minttypes.ParamKeyTable()
			case alliancetypes.ModuleName:
				keyTable = alliancetypes.ParamKeyTable()
			}

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		// Mint module params update
		params := minttypes.Params{}
		params.BlocksPerYear = 5733818
		params.TotalBurntAmount = []sdk.Coin{sdk.NewInt64Coin("ufury", 20_000_000_000000)}
		subspace, ok := keepers.ParamsKeeper.GetSubspace(minttypes.ModuleName)
		if !ok {
			panic("invalid mint module subspace")
		}
		v := reflect.Indirect(reflect.ValueOf(params.BlocksPerYear)).Interface()
		subspace.Set(ctx, minttypes.KeyBlocksPerYear, v)
		v = reflect.Indirect(reflect.ValueOf(params.TotalBurntAmount)).Interface()
		subspace.Set(ctx, minttypes.KeyTotalBurntAmount, v)

		cp := tmtypes.DefaultConsensusParams().ToProto()
		keepers.ConsensusParamsKeeper.Set(ctx, &tmproto.ConsensusParams{
			Block:     cp.Block,
			Validator: cp.Validator,
			Evidence:  cp.Evidence,
			Version:   cp.Version,
		})

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
