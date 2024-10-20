package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk-starter-kit/consts"
	mconsts "github.com/ava-labs/hypersdk-starter-kit/consts"
	"github.com/ava-labs/hypersdk-starter-kit/storage"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"
)

const (
	AssetTransferComputeUnits = 1
	MaxReasonSize             = 256
)

var (
	ErrReasonTooLarge              = errors.New("reason is too large")
	ErrAssetNotOwned               = errors.New("asset not owned")
	_                 chain.Action = (*AssetTransfer)(nil)
)

type AssetTransfer struct {
	// Recepient is the recipient of the [Asset].
	Recipient codec.Address `serialize:"true" json:"to"`

	// Asset  is transferred to [To].
	Asset ids.ID `serialize:"true" json:"asset"`

	// Reason for transfer.
	Reason string `serialize:"true" json:"reason"`
}

// GetTypeID implements chain.Action.
func (a *AssetTransfer) GetTypeID() uint8 {
	return consts.AssetTransferID
}

// StateKeys implements chain.Action.
func (a *AssetTransfer) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.AssetKey(a.Asset)): state.All,
	}
	// Here we are not interested on keys from the actor
}

var _ codec.Typed = (*AssetTransferResult)(nil)

type AssetTransferResult struct {
	OldOwner codec.Address `serialize:"true" json:"old_owner"`
	NewOwner codec.Address `serialize:"true" json:"new_owner"`
}

func (*AssetTransferResult) GetTypeID() uint8 {
	return mconsts.AssetTransferID // Common practice is to use the action ID
}

// Execute implements chain.Action.
func (a *AssetTransfer) Execute(
	ctx context.Context,
	r chain.Rules,
	mu state.Mutable,
	timestamp int64,
	actor codec.Address,
	actionID ids.ID,
) (codec.Typed, error) {
	if len(a.Reason) > MaxReasonSize {
		return nil, ErrReasonTooLarge
	}
	oldOwner, err := storage.GetAssetOwner(ctx, mu, a.Asset)
	if err != nil {
		return nil, err
	}
	if oldOwner != actor {
		return nil, ErrAssetNotOwned
	}
	err = storage.ChangeAssetOwner(ctx, mu, a.Asset, a.Recipient)
	if err != nil {
		return nil, err
	}
	return &AssetTransferResult{
		OldOwner: oldOwner,
		NewOwner: a.Recipient,
	}, nil
}

// ComputeUnits implements chain.Action.
func (a *AssetTransfer) ComputeUnits(chain.Rules) uint64 {
	return AssetTransferComputeUnits
}

// ValidRange implements chain.Action.
func (a *AssetTransfer) ValidRange(chain.Rules) (start int64, end int64) {
	return -1, -1
}
