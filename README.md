# HyperSDK Starter Kit

HyperSDK Starter includes:
- Boilerplate VM based on [MorpheusVM](https://github.com/ava-labs/hypersdk/tree/main/examples/morpheusvm)
- Universal frontend
- Metamask Snap wallet
- A quick start guide (this document)

## Prerequisites
- Docker (recent version)
- Optional: Golang v1.22.5+
- Optional: NodeJS v20+
- Optional: [Metamask Flask](https://chromewebstore.google.com/detail/metamask-flask-developmen/ljfoeinjpaedjfecbmggjgodbgkmjkjk). Disable normal Metamask, Core wallet, and other wallets. *Do not use your real private key with Flask*.

## 0. Clone this repo
`git clone https://github.com/ava-labs/hypersdk-starter-kit.git`

## 1. Launch this example

Run: `docker compose up -d --build devnet faucet frontend`. This may take 5 minutes to download dependencies.

For devcontainers or codespaces, forward ports `8765` for faucet, `9650` for the chain, and `5173` for the frontend.

When finished, stop everything with: `docker compose down`

## 2. Explore MorpheusVM
This repo includes [MorpheusVM](https://github.com/ava-labs/hypersdk/tree/main/examples/morpheusvm), the simplest HyperSDK VM. It supports one action (Transfer) for moving funds and tracking balances.

### 2.1 Connect wallet
Open [http://localhost:5173](http://localhost:5173) to see the frontend.

![Auth options](assets/auth.png)

We recommend using a Snap (requires [Metamask Flask](https://chromewebstore.google.com/detail/metamask-flask-developmen/ljfoeinjpaedjfecbmggjgodbgkmjkjk)) for the full experience, but a temporary wallet works too.

### 2.2 Execute a read-only action

Actions can be executed on-chain (in a transaction) with results saved to a block, or off-chain (read-only). MorpheusVM has one action. Try executing it read-only. It shows expected balances of the sender and receiver. See the logic in `actions/transfer.go`.

![Read-only action](assets/read-only.png)

### 2.3 Issue a transaction

Now, write data to the chain. Click "Execute in transaction". All fields are pre-filled with default values.

![Sign](assets/sign.png)

After mining, the transaction appears in the right column. This column shows all non-empty blocks on the chain.

### 2.4 Check Logs

Logs are located inside the Docker container. To view them, you'll need to open a bash terminal inside the container and navigate to the folder with the current network:
```bash
docker exec -it devnet bash -c "cd /root/.tmpnet/networks/latest_morpheusvm-e2e-tests && bash"
```

This isn’t the best developer experience, and we’re working on improving it.

## 3. Add Your Own Custom Action

Think of actions in HyperSDK like functions in EVMs. They have inputs, outputs, and execution logic.

Let's add the `Greeting` action. This action doesn’t change anything; it simply prints your balance and the current date. However, if it's executed in a transaction, the output will be recorded in a block on the chain.

### 3.1 Create an Action File

Place the following code in `actions/greeting.go`. The code includes some comments, but for more details, check out [the docs folder in HyperSDK](https://github.com/ava-labs/hypersdk/tree/main/docs).
```golang
package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk-starter-kit/consts"
	"github.com/ava-labs/hypersdk-starter-kit/storage"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"
)

// Please see chain.Action interface description for more information
var _ chain.Action = (*Greeting)(nil)

// Action struct. All "serialize" marked fields will be saved on chain
type Greeting struct {
	Name string `serialize:"true" json:"name"`
}

// TypeID, has to be unique across all actions
func (*Greeting) GetTypeID() uint8 {
	return consts.HiID
}

// All database keys that could be touched during execution.
// Will fail if a key is missing or has wrong permissions
func (g *Greeting) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.BalanceKey(actor)): state.Read,
	}
}

// The "main" function of the action
func (g *Greeting) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable, // That's how we read and write to the database
	timestamp int64, // Timestamp of the block or the time of simulation
	actor codec.Address, // Whoever signed the transaction, or a placeholder address in case of read-only action
	_ ids.ID, // actionID
) (codec.Typed, error) {
	balance, err := storage.GetBalance(ctx, mu, actor)
	if err != nil {
		return nil, err
	}
	currentTime := time.Unix(timestamp/1000, 0).Format("January 2, 2006")
	greeting := fmt.Sprintf(
		"Hi, dear %s! Today, %s, your balance is %s %s.",
		g.Name,
		currentTime,
		utils.FormatBalance(balance),
		consts.Symbol,
	)

	return &GreetingResult{
		Greeting: greeting,
	}, nil
}

// How many compute units to charge for executing this action. Can be dynamic based on the action.
func (*Greeting) ComputeUnits(chain.Rules) uint64 {
	return 1
}

// ValidRange is the timestamp range (in ms) that this [Action] is considered valid.
// -1 means no start/end
func (*Greeting) ValidRange(chain.Rules) (int64, int64) {
	return -1, -1
}

// Result of execution of greeting action
type GreetingResult struct {
	Greeting string `serialize:"true" json:"greeting"`
}

// Has to implement codec.Typed for on-chain serialization
var _ codec.Typed = (*GreetingResult)(nil)

// TypeID of the action result, could be the same as the action ID
func (g *GreetingResult) GetTypeID() uint8 {
	return consts.HiID
}
```
### 3.2 Register the Action

Now, you need to make both the VM and clients (via ABI) aware of this action.

To do this, register your action in `vm/vm.go` after the line `ActionParser.Register(&actions.Transfer{}, nil):`
```golang
ActionParser.Register(&actions.Greeting{}, nil),
```

Then, register its output after the line `OutputParser.Register(&actions.TransferResult{}, nil):`
```golang
OutputParser.Register(&actions.GreetingResult{}, nil),
```

### 3.3 Rebuild Your VM
```bash
docker compose down -t 1; docker compose up -d --build devnet faucet frontend
```

### 3.4 Test Your New Action

HyperSDK uses ABI, an autogenerated description of all the actions in your VM. Thanks to this, the frontend already knows how to interact with your new action. Every action you add will be displayed on the frontend and supported by the wallet as soon as the node restarts.

Now, enter your name and see the result:

![Greeting result](assets/greeting.png)

You can also send it as a transaction, but this doesn't make much sense since there’s nothing to write to the chain's state.

### 3.5 Next Steps

Congrats! You've just created your first action for HyperSDK.

This covers nearly half of what you need to build your own blockchain on HyperSDK. The remaining part is state management, which you can explore in `storage/storage.go`. Dive in and enjoy your journey!

## 4. Develop a Frontend
1. If you started anything, bring everything down: `docker compose down`
2. Start only the devnet and faucet: `docker compose up -d --build devnet faucet`
3. Navigate to the web wallet: `cd web_wallet`
4. Install dependencies and start the dev server: `npm i && npm run dev`

Make sure ports `8765` (faucet), `9650` (chain), and `5173` (frontend) are forwarded.

Learn more from [npm:hypersdk-client](https://www.npmjs.com/package/hypersdk-client) and the `web_wallet` folder in this repo.

## 5. Playing with storage

The HyperSDK offers vast opportunities to build custom storage mechanisms for your virtual machine (VM). By customizing how assets are stored in your VM, you can tailor the experience to meet the needs of your specific use case.

### Why Customize Storage?

With HyperSDK, customizing storage gives you complete control over how data is handled on-chain. Whether you’re building a complex financial system, creating dynamic rules for asset ownership, or experimenting with novel consensus mechanisms, **storage customization empowers you to align the chain's functionality with your vision**.

### Tutorial: Customizing storage for Real World Assets in HyperSDK
This tutorial will guide you through the steps to leverage a newly implemented asset storage in HyperSDK, highlighting the incredible flexibility it brings to your VM development.

 In this tutorial, we’ll demonstrate how to set up a custom storage mechanism for seamless and efficient transferring of assets, focusing on how to:
- Define unique keys for asset ownership.
- Efficiently query and modify asset owners.
- Ensure secure and atomic state updates.

#### Step 1: Taking a look at the storage layout MorpheusVM presents

When we first open the `storage.go` file, we'll find some constants representing prefixes. It's with the use of these prefixes that we're going to be able to separate different kinds of values being stored. 

Up until now, we just have metadata required by the HyperSDK and a mapping that has accounts as keys and their balances as values.

But what if I now want to represent some specific type of asset, which is to be jut owned by some address. Let's take a dive into that scenario.


#### Step 2: Adding a new partition to the storage

The idea is that now we'll have another mapping in our storage. One which has _Assets_ as keys and _Addresses_ as values.

Firstly, what we should do in order to add this mapping would be to create a new prefix. So the first part of `storage.go` should look like this

```go
const (
	// Active state
	balancePrefix   = 0x0
	heightPrefix    = 0x1
	timestampPrefix = 0x2
	feePrefix       = 0x3
	assetPrefix     = 0x4
)
```

And secondly, we'll have to define the AssetKey function, in order to, given an asset id, be able to return the state key that stores the asset's owner.

```go
const AssetChunks uint16 = 1

func AssetKey(assetID ids.ID) (k []byte) {
	k = make([]byte, 1+ids.IDLen+consts.Uint16Len)
	k[0] = assetPrefix
	copy(k[1:], assetID[:])
	binary.BigEndian.PutUint16(k[1+ids.IDLen:], AssetChunks)
	return
}
```

Before the actual declaration of the functions, readers could ask what is the AssentChunks constant and what does it have to do with anything. 

The HyperSDK requires using size-encoded storage keys, which include a "chunk suffix." The "chunk suffix" is encoded as a uint16. The length encoded in the "chunk suffix" sets the maximum length of any value that can ever be placed in the corresponding key-value pair. This also means trying to change the suffix will change the key itself. In other words, if you write to the same key with a different suffix, it will write to a different location instead of overwriting. This means that the suffix must provide an upper bound for the value size forever or the VM would need to explicitly handle migrating from the key from size A to size B.

Chunks are given in batches of 64 bytes, so a chunk size of 1 means the value at that key will never exceed 64 bytes (2 -> max of 128 bytes, 3 -> max of 192, etc).

Declaring AssetChunks as 1, means we're committing to always store a value of less than 64 bytes (which makes sense since we're storing addresses as values).


**Technical note**: Why would I want to store my asset data like this? This precise example comes from the motivation of tokenizing unique real world assets, such as real state. See this document for more information.

#### Step 3: Handling this new partition

Now, we're able to implement different functions that interact with this new component of the state. Here we present some helper functions to get and set the asset's owner:

```go

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
```

#### Step 4: What am I able to do now that a have a new storage?

It might (surely) be cliche to say the sky's the limit, but already knowing that you can create your own actions on the VM, you know at least partially true.

However, for this tutorial, we'll just implement a new action that interacts with this new component of the state.

In the `actions` folder we should create a new file `transfer_asset.go` containing the following:

```go
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
```

And a small change should be done on the `consts/types.go` file, adding the AssetTransferID declaration:

```go
const (
	// Action TypeIDs
	TransferID      uint8 = 0
	AssetTransferID uint8 = 1
)
```

Now, we're able to change ownerships of assets. Now it's your turn to create all state partitions and actions you can imagine. Just think and play :D

### Unlocking New Possibilities with Custom Storage

As stated, this is just the beginning of what’s possible when you take full control of your VM's storage. 

By leveraging HyperSDK’s state module and customizing it to your requirements, you can build innovative, secure, and scalable decentralized applications (dApps). **Your storage layer becomes a powerful tool, driving the features that set your VM apart**.

## Notes
- You can launch everything without Docker:
  - Faucet: `go run ./cmd/faucet/`
  - Chain: `./scripts/run.sh`, and use `./scripts/stop.sh` to stop
  - Frontend: `npm run dev` in `web_wallet`
- Be aware of potential port conflicts. If issues arise, `docker rm -f $(docker ps -a -q)` will help.
- For VM development, you don’t need to know JavaScript—you can use an existing frontend, and all actions will be added automatically.
- If the frontend works with an ephemeral private key but doesn't work with the Snap, delete the Snap, refresh the page, and try again. The Snap might be outdated.
- Instead of using `./build/morpheus-cli` commands, please directly use `go run ./cmd/morpheus-cli/` for the CLI.
- Always ensure that you have the `hypersdk-client` npm version and the golang `github.com/ava-labs/hypersdk` version from the same commit of the starter kit. HyperSDK evolves rapidly.
