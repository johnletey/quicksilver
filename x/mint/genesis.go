package mint

import (
	"github.com/ingenuity-build/quicksilver/x/mint/keeper"
	"github.com/ingenuity-build/quicksilver/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis new mint genesis.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, ak types.AccountKeeper, data *types.GenesisState) {
	data.Minter.EpochProvisions = data.Params.GenesisEpochProvisions
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)

	if !ak.HasAccount(ctx, ak.GetModuleAddress(types.ModuleName)) {
		ak.GetModuleAccount(ctx, types.ModuleName)
	}

	keeper.SetLastReductionEpochNum(ctx, data.ReductionStartedEpoch)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	lastReductionEpoch := keeper.GetLastReductionEpochNum(ctx)
	return types.NewGenesis(minter, params, lastReductionEpoch)
}
