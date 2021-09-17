package wallet

import (
	"fmt"

	pb "github.com/ariden83/bitcoin-wallet/proto/btchdwallet"

	"github.com/ariden83/bitcoin-wallet/config"
	"github.com/ariden83/bitcoin-wallet/crypt"
	"github.com/blockcypher/gobcy"
	"github.com/brianium/mnemonic"
	"github.com/wemeetagain/go-hdwallet"
	"go.uber.org/zap"
)

type Wallet struct {
	log  *zap.Logger
	conf *config.Config
}

func New(conf *config.Config, l *zap.Logger) *Wallet {
	return &Wallet{
		conf: conf,
		log:  l,
	}
}

// CreateWallet is in charge of creating a new root wallet
func (Wallet) CreateWallet() *pb.Response {
	// Generate a random 256 bit seed
	seed := crypt.CreateHash()
	mnemonic, _ := mnemonic.New([]byte(seed), mnemonic.English)

	// Create a master private key
	masterprv := hdwallet.MasterKey([]byte(mnemonic.Sentence()))

	// Convert a private key to public key
	masterpub := masterprv.Pub()

	// Get your address
	address := masterpub.Address()

	return &pb.Response{Address: address, PubKey: masterpub.String(), PrivKey: masterprv.String(), Mnemonic: mnemonic.Sentence()}
}

// DecodeWallet is in charge of decoding wallet from mnemonic
func (Wallet) DecodeWallet(mnemonic string) *pb.Response {
	// Get private key from mnemonic
	masterprv := hdwallet.MasterKey([]byte(mnemonic))

	// Convert a private key to public key
	masterpub := masterprv.Pub()

	// Get your address
	address := masterpub.Address()

	return &pb.Response{Address: address, PubKey: masterpub.String(), PrivKey: masterprv.String()}
}

// GetBalance is in charge of returning the given address balance
func (w *Wallet) GetBalance(address string) *pb.Response {
	btc := gobcy.API{w.conf.BlockCypher.Token, "btc", "main"}
	addr, err := btc.GetAddrBal(address, nil)
	if err != nil {
		fmt.Println(err)
	}

	balance := addr.Balance
	totalReceived := addr.TotalReceived
	totalSent := addr.TotalSent
	unconfirmedBalance := addr.UnconfirmedBalance

	return &pb.Response{
		Address:            address,
		Balance:            int64(balance.Uint64()),
		TotalReceived:      int64(totalReceived.Uint64()),
		TotalSent:          int64(totalSent.Uint64()),
		UnconfirmedBalance: int64(unconfirmedBalance.Uint64()),
	}
}
