package balance_test

import (
	"encoding/base64"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/s7techlab/cckit/router"
	testcc "github.com/s7techlab/cckit/testing"

	"github.com/s7techlab/cckit/identity"
	"github.com/s7techlab/cckit/identity/testdata"

	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/balance"
)

func TestBalance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Balance test suite")
}

var (
	ownerIdentity = testdata.Certificates[0].MustIdentity(testdata.DefaultMSP)
	user1Identity = testdata.Certificates[1].MustIdentity(testdata.DefaultMSP)
	user2Identity = testdata.Certificates[2].MustIdentity(testdata.DefaultMSP)

	ownerAddress = base64.StdEncoding.EncodeToString(identity.MarshalPublicKey(ownerIdentity.Cert.PublicKey))
	user1Address = base64.StdEncoding.EncodeToString(identity.MarshalPublicKey(user1Identity.Cert.PublicKey))
	user2Address = base64.StdEncoding.EncodeToString(identity.MarshalPublicKey(user2Identity.Cert.PublicKey))

	Symbol = `AA`

	TotalSupply = uint64(10000000)
)

type Wallet struct {
	cc      *testcc.TxHandler
	ctx     router.Context     // wallet storage here
	store   *balance.UTXOStore //  balance access
	address string
	symbol  string
}

func NewWallet(cc *testcc.TxHandler, ctx router.Context, store *balance.UTXOStore, address, symbol string) *Wallet {
	return &Wallet{
		cc:      cc,
		ctx:     ctx,
		store:   store,
		address: address,
		symbol:  symbol,
	}
}

func (w *Wallet) ExpectBalance(amount uint64) {
	b, err := w.store.Get(w.ctx, &balance.BalanceId{
		Address: w.address,
		Symbol:  w.symbol,
	})

	Expect(err).NotTo(HaveOccurred())
	Expect(b.Amount).To(Equal(amount))
}

func (w *Wallet) ExpectMint(amount uint64) {
	w.cc.Tx(func() {
		err := w.store.Mint(w.ctx, &balance.BalanceOperation{
			Address: w.address,
			Symbol:  w.symbol,
			Amount:  amount,
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectBurn(amount uint64) {
	w.cc.Tx(func() {
		err := w.store.Burn(w.ctx, &balance.BalanceOperation{
			Address: w.address,
			Symbol:  w.symbol,
			Amount:  amount,
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectTransfer(recipient string, amount uint64) {
	w.cc.Tx(func() {
		err := w.store.Transfer(w.ctx, &balance.TransferOperation{
			Sender:    w.address,
			Recipient: recipient,
			Symbol:    w.symbol,
			Amount:    amount,
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

type transfer struct {
	recipient string
	amount    uint64
}

func (w *Wallet) ExpectTransferBatch(transfers []*transfer) {

	var transferOperations []*balance.TransferOperation

	for _, t := range transfers {
		transferOperations = append(transferOperations, &balance.TransferOperation{
			Sender:    w.address,
			Recipient: t.recipient,
			Symbol:    w.symbol,
			Amount:    t.amount,
		})
	}
	w.cc.Tx(func() {
		err := w.store.TransferBatch(w.ctx, transferOperations)
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectOutputsNum(num int) {
	outputs, err := w.store.ListOutputs(w.ctx, &balance.BalanceId{
		Address: w.address,
		Symbol:  w.symbol,
	})

	Expect(err).NotTo(HaveOccurred())
	Expect(len(outputs)).To(Equal(num))
}

var _ = Describe(`UTXO store`, func() {

	cc, ctx := testcc.NewTxHandler(`UTXO`)
	utxo := balance.NewUTXOStore()
	ownerWallet := NewWallet(cc, ctx, utxo, ownerAddress, Symbol)
	user1Wallet := NewWallet(cc, ctx, utxo, user1Address, Symbol)
	user2Wallet := NewWallet(cc, ctx, utxo, user2Address, Symbol)

	It(`allow to ge empty balance`, func() {
		ownerWallet.ExpectBalance(0)
	})

	It(`allow to mint balance`, func() {
		ownerWallet.ExpectMint(TotalSupply)
		ownerWallet.ExpectBalance(TotalSupply)
		ownerWallet.ExpectOutputsNum(1)
	})

	It(`allow to mint balance once more time`, func() {
		ownerWallet.ExpectMint(TotalSupply)
		ownerWallet.ExpectBalance(TotalSupply * 2)
		ownerWallet.ExpectOutputsNum(2)
	})

	It(`allow to partially transfer balance`, func() {
		ownerWallet.ExpectTransfer(user1Address, 100)
		ownerWallet.ExpectBalance(TotalSupply*2 - 100)
		ownerWallet.ExpectOutputsNum(2)

		user1Wallet.ExpectBalance(100)
		user1Wallet.ExpectOutputsNum(1)
	})

	It(`allow to return all amount back`, func() {
		user1Wallet.ExpectTransfer(ownerAddress, 100)
		ownerWallet.ExpectBalance(TotalSupply * 2)
		ownerWallet.ExpectOutputsNum(3)

		user1Wallet.ExpectBalance(0)
		user1Wallet.ExpectOutputsNum(0)
	})

	It(`allow to burn`, func() {
		ownerWallet.ExpectBurn(TotalSupply)
		ownerWallet.ExpectBalance(TotalSupply)
		//ownerWallet.ExpectOutputsNum(2)
	})

	It(`allow to trasfer batch`, func() {
		ownerWallet.ExpectTransferBatch([]*transfer{
			{recipient: user1Address, amount: 100},
			{recipient: user2Address, amount: 200},
		})
		ownerWallet.ExpectBalance(TotalSupply - 100 - 200)
		user1Wallet.ExpectBalance(100)
		user2Wallet.ExpectBalance(200)
		//ownerWallet.ExpectOutputsNum(2)
	})
})
