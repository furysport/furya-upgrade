package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/furysport/furya-upgrade/x/airdrop/types"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper struct
type Keeper struct {
	cdc           codec.Codec
	storeKey      storetypes.StoreKey
	paramSpace    paramstypes.Subspace
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	acountKeeper  types.AccountKeeper
}

// NewKeeper returns keeper
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	paramSpace paramstypes.Subspace,
	bk types.BankKeeper, sk types.StakingKeeper, ak types.AccountKeeper) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramSpace,
		bankKeeper:    bk,
		stakingKeeper: sk,
		acountKeeper:  ak,
	}
}

// Logger returns logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
