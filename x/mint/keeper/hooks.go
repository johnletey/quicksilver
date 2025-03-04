package keeper

import (
	"fmt"

	epochstypes "github.com/ingenuity-build/quicksilver/x/epochs/types"
	"github.com/ingenuity-build/quicksilver/x/mint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BeforeEpochStart(_ sdk.Context, _ string, _ int64) error {
	return nil
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	params := k.GetParams(ctx)
	k.Logger(ctx).Info("Mint AfterEpochEnd", "Params", params)

	if epochIdentifier == params.EpochIdentifier {
		// not distribute rewards if it's not time yet for rewards distribution
		if epochNumber < params.MintingRewardsDistributionStartEpoch {
			return nil
		} else if epochNumber == params.MintingRewardsDistributionStartEpoch {
			k.SetLastReductionEpochNum(ctx, epochNumber)
		}
		// fetch stored minter & params
		minter := k.GetMinter(ctx)
		params := k.GetParams(ctx)

		// Check if we have hit an epoch where we update the inflation parameter.
		// Since epochs only update based on BFT time data, it is safe to store the "reductioning period time"
		// in terms of the number of epochs that have transpired.
		if epochNumber >= k.GetParams(ctx).ReductionPeriodInEpochs+k.GetLastReductionEpochNum(ctx) {
			// reduction the reward per reduction period
			minter.EpochProvisions = minter.NextEpochProvisions(params)
			k.SetMinter(ctx, minter)
			k.SetLastReductionEpochNum(ctx, epochNumber)
		}

		// mint coins, update supply
		mintedCoin := minter.EpochProvision(params)
		mintedCoins := sdk.NewCoins(mintedCoin)

		// We over-allocate by the developer vesting portion, and burn this later
		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			return err
		}

		// send the minted coins to the fee collector account
		err = k.DistributeMintedCoin(ctx, mintedCoin)
		if err != nil {
			return err
		}

		if mintedCoin.Amount.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMint,
				sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochNumber)),
				sdk.NewAttribute(types.AttributeKeyEpochProvisions, minter.EpochProvisions.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
			),
		)
	}
	return nil
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
