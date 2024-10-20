// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
)

type ReadState func(context.Context, [][]byte) ([][]byte, []error)

// State
// / (height) => store in root
//   -> [heightPrefix] => height
// 0x0/ (balance)
//   -> [owner] => balance
// 0x1/ (hypersdk-height)
// 0x2/ (hypersdk-timestamp)
// 0x3/ (hypersdk-fee)
// 0x4/ (hypersdk-asset)
//   -> [assetID] => owner

const (
	// Active state
	balancePrefix   = 0x0
	heightPrefix    = 0x1
	timestampPrefix = 0x2
	feePrefix       = 0x3
	assetPrefix     = 0x4
)

const BalanceChunks uint16 = 1
const AssetChunks uint16 = 1

var (
	heightKey    = []byte{heightPrefix}
	timestampKey = []byte{timestampPrefix}
	feeKey       = []byte{feePrefix}
)

// we're using ids.ID as the key for assets but might want to switch to an
// specific data type
// [assetPrefix] + [assetID]
func AssetKey(assetID ids.ID) (k []byte) {
	k = make([]byte, 1+ids.IDLen+consts.Uint16Len)
	k[0] = assetPrefix
	copy(k[1:], assetID[:])
	binary.BigEndian.PutUint16(k[1+ids.IDLen:], AssetChunks)
	return
}

func GetAssetOwner(
	ctx context.Context,
	im state.Immutable,
	assetID ids.ID,
) (codec.Address, error) {
	_, owner, _, err := getAssetOwner(ctx, im, assetID)
	return owner, err
}

func getAssetOwner(
	ctx context.Context,
	im state.Immutable,
	assetID ids.ID,
) ([]byte, codec.Address, bool, error) {
	k := AssetKey(assetID)
	owner, exists, err := innerGetAssetOwner(im.GetValue(ctx, k))
	return k, owner, exists, err
}

func innerGetAssetOwner(
	v []byte,
	err error,
) (codec.Address, bool, error) {
	if errors.Is(err, database.ErrNotFound) {
		return codec.EmptyAddress, false, nil
	}
	if err != nil {
		return codec.EmptyAddress, false, err
	}
	val, err := codec.ToAddress(v)
	if err != nil {
		return codec.EmptyAddress, false, err
	}
	return val, true, nil
}

func GetAssetOwnerFromState(
	ctx context.Context,
	f ReadState,
	assetID ids.ID,
) (codec.Address, error) {
	k := AssetKey(assetID)
	values, errs := f(ctx, [][]byte{k})
	owner, _, err := innerGetAssetOwner(values[0], errs[0])
	return owner, err
}

func SetAssetOwner(
	ctx context.Context,
	mu state.Mutable,
	key []byte,
	newowner codec.Address,
) error {
	byteNewOwner, err := newowner.MarshalText()
	if err != nil {
		return mu.Insert(ctx, key, byteNewOwner)
	}
	return err
}

func ChangeAssetOwner(
	ctx context.Context,
	mu state.Mutable,
	assetID ids.ID,
	newOwner codec.Address,
) error {
	k := AssetKey(assetID)
	return SetAssetOwner(ctx, mu, k, newOwner)
}

// [balancePrefix] + [address]
func BalanceKey(addr codec.Address) (k []byte) {
	k = make([]byte, 1+codec.AddressLen+consts.Uint16Len)
	k[0] = balancePrefix
	copy(k[1:], addr[:])
	binary.BigEndian.PutUint16(k[1+codec.AddressLen:], BalanceChunks)
	return
}

// If locked is 0, then account does not exist
func GetBalance(
	ctx context.Context,
	im state.Immutable,
	addr codec.Address,
) (uint64, error) {
	_, bal, _, err := getBalance(ctx, im, addr)
	return bal, err
}

func getBalance(
	ctx context.Context,
	im state.Immutable,
	addr codec.Address,
) ([]byte, uint64, bool, error) {
	k := BalanceKey(addr)
	bal, exists, err := innerGetBalance(im.GetValue(ctx, k))
	return k, bal, exists, err
}

// Used to serve RPC queries
func GetBalanceFromState(
	ctx context.Context,
	f ReadState,
	addr codec.Address,
) (uint64, error) {
	k := BalanceKey(addr)
	values, errs := f(ctx, [][]byte{k})
	bal, _, err := innerGetBalance(values[0], errs[0])
	return bal, err
}

func innerGetBalance(
	v []byte,
	err error,
) (uint64, bool, error) {
	if errors.Is(err, database.ErrNotFound) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	val, err := database.ParseUInt64(v)
	if err != nil {
		return 0, false, err
	}
	return val, true, nil
}

func SetBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	balance uint64,
) error {
	k := BalanceKey(addr)
	return setBalance(ctx, mu, k, balance)
}

func setBalance(
	ctx context.Context,
	mu state.Mutable,
	key []byte,
	balance uint64,
) error {
	return mu.Insert(ctx, key, binary.BigEndian.AppendUint64(nil, balance))
}

func AddBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	amount uint64,
	create bool,
) (uint64, error) {
	key, bal, exists, err := getBalance(ctx, mu, addr)
	if err != nil {
		return 0, err
	}
	// Don't add balance if account doesn't exist. This
	// can be useful when processing fee refunds.
	if !exists && !create {
		return 0, nil
	}
	nbal, err := smath.Add(bal, amount)
	if err != nil {
		return 0, fmt.Errorf(
			"%w: could not add balance (bal=%d, addr=%v, amount=%d)",
			ErrInvalidBalance,
			bal,
			addr,
			amount,
		)
	}
	return nbal, setBalance(ctx, mu, key, nbal)
}

func SubBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	amount uint64,
) (uint64, error) {
	key, bal, ok, err := getBalance(ctx, mu, addr)
	if !ok {
		return 0, ErrInvalidAddress
	}
	if err != nil {
		return 0, err
	}
	nbal, err := smath.Sub(bal, amount)
	if err != nil {
		return 0, fmt.Errorf(
			"%w: could not subtract balance (bal=%d, addr=%v, amount=%d)",
			ErrInvalidBalance,
			bal,
			addr,
			amount,
		)
	}
	if nbal == 0 {
		// If there is no balance left, we should delete the record instead of
		// setting it to 0.
		return 0, mu.Remove(ctx, key)
	}
	return nbal, setBalance(ctx, mu, key, nbal)
}

func HeightKey() (k []byte) {
	return heightKey
}

func TimestampKey() (k []byte) {
	return timestampKey
}

func FeeKey() (k []byte) {
	return feeKey
}
