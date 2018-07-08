package paychan

// keeper of the paychan store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *wire.Codec // needed?
	coinKeeper bank.Keeper

	// codespace
	codespace sdk.CodespaceType // ??
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		coinKeeper: ck,
		codespace:  codespace,
	}
	return keeper
}

// bunch of business logic ...

func (keeper Keeper) GetPaychan(paychanID) Paychan {
	// load from DB
	// unmarshall
	// return
}

func (keeper Keeper) setPaychan(pych Paychan) sdk.Error {
	// marshal
	// write to db
}

func (keeer Keeper) CreatePaychan(receiver sdkAddress, amt sdk.Coins) (Paychan, sdk.Error) {
	// subtract coins from sender
	// create new Paychan struct (create ID)
	// save to db

	// validation:
	// sender has enough coins
	// receiver address exists?
	// paychan doesn't exist already
}

func (keeper Keeper) ClosePaychan() sdk.Error {
	// add coins to sender
	// add coins to receiver
	// delete paychan from db

	// validation:
	// paychan exists
	// output coins are less than paychan balance
	// sender and receiver addresses exist?
}

func paychanKey(Paychan) {
	// concat sender and receiver and integer ID
}

// maybe getAllPaychans(sender sdk.address) []Paychan
