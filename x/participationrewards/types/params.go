package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var (
	KeyDistributionProportions = []byte("DistributionProportions")
	KeyClaimsEnabled           = []byte("ClaimsEnabled")

	DefaultValidatorSelectionAllocation = sdk.NewDecWithPrec(34, 2)
	DefaultHoldingsAllocation           = sdk.NewDecWithPrec(33, 2)
	DefaultLockupAllocation             = sdk.NewDecWithPrec(33, 2)
	DefaultClaimsEnabled                = false
)

// ParamTable for participationrewards module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new ics Params instance
func NewParams(
	validatorSelectionAllocation sdk.Dec,
	holdingsAllocation sdk.Dec,
	lockupAllocation sdk.Dec,
	claimsEnabled bool,
) Params {
	return Params{
		DistributionProportions: DistributionProportions{
			ValidatorSelectionAllocation: validatorSelectionAllocation,
			HoldingsAllocation:           holdingsAllocation,
			LockupAllocation:             lockupAllocation,
		},
		ClaimsEnabled: claimsEnabled,
	}
}

// DefaultParams default ics params
func DefaultParams() Params {
	return NewParams(
		DefaultValidatorSelectionAllocation,
		DefaultHoldingsAllocation,
		DefaultLockupAllocation,
		DefaultClaimsEnabled,
	)
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistributionProportions, &p.DistributionProportions, validateDistributionProportions),
		paramtypes.NewParamSetPair(KeyClaimsEnabled, &p.ClaimsEnabled, validateBoolean),
	}
}

// ParamSetPairs implements params.ParamSet
func (p *ParamsV1) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistributionProportions, &p.DistributionProportions, validateDistributionProportions),
	}
}

func validateDistributionProportions(i interface{}) error {
	dp, ok := i.(DistributionProportions)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return dp.ValidateBasic()
}

func validateBoolean(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// validate params.
func (p Params) Validate() error {
	return validateDistributionProportions(p.DistributionProportions)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// String implements the Stringer interface.
func (p ParamsV1) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
