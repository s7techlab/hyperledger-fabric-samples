package erc20_test

import (
	"encoding/base64"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/s7techlab/cckit/identity"
	"github.com/s7techlab/cckit/identity/testdata"
	testcc "github.com/s7techlab/cckit/testing"
	expectcc "github.com/s7techlab/cckit/testing/expect"

	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/chaincode/erc20"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/account"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/allowance"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/balance"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/burnable"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/config_erc20"
)

func TestERC20(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ERC20 Test suite")
}

var (
	ownerIdentity = testdata.Certificates[0].MustIdentity(testdata.DefaultMSP)
	user1Identity = testdata.Certificates[1].MustIdentity(testdata.DefaultMSP)
	user2Identity = testdata.Certificates[2].MustIdentity(testdata.DefaultMSP)

	ownerAddress = base64.StdEncoding.EncodeToString(identity.MarshalPublicKey(ownerIdentity.Cert.PublicKey))
	user1Address = base64.StdEncoding.EncodeToString(identity.MarshalPublicKey(user1Identity.Cert.PublicKey))
	user2Address = base64.StdEncoding.EncodeToString(identity.MarshalPublicKey(user2Identity.Cert.PublicKey))
)

func init() {

	var stores = []balance.Store{
		balance.NewAccountStore(), balance.NewUTXOStore(),
	}

	for _, store := range stores {

		var (
			name = `ERC-20 ` + reflect.TypeOf(store).String()
			cc   *testcc.MockStub
		)

		var _ = Describe(name, func() {

			Context(`init`, func() {

				It(`Allow to  init`, func() {
					chaincode, err := erc20.New(name, store)
					Expect(err).NotTo(HaveOccurred())
					cc = testcc.NewMockStub(name, chaincode)

					expectcc.ResponseOk(cc.From(ownerIdentity).Init())
				})

				It(`Allow to call init once more time `, func() {
					expectcc.ResponseOk(cc.From(ownerIdentity).Init())
				})

			})

			Context(`token info`, func() {

				It(`Allow to get token name`, func() {
					name := expectcc.PayloadIs(
						cc.From(user1Identity).
							Query(config_erc20.ConfigERC20ServiceChaincode_GetName, nil),
						&config_erc20.NameResponse{}).(*config_erc20.NameResponse)

					Expect(name.Name).To(Equal(erc20.Token.Name))
				})
			})

			Context(`initial balance`, func() {

				It(`Allow to know invoker address `, func() {
					address := expectcc.PayloadIs(
						cc.From(user1Identity).
							Query(account.AccountServiceChaincode_GetInvokerAddress, nil),
						&account.AddressId{}).(*account.AddressId)

					Expect(address.Address).To(Equal(user1Address))

					address = expectcc.PayloadIs(
						cc.From(user2Identity).
							Query(account.AccountServiceChaincode_GetInvokerAddress, nil),
						&account.AddressId{}).(*account.AddressId)

					Expect(address.Address).To(Equal(user2Address))
				})

				It(`Allow to get owner balance`, func() {
					b := expectcc.PayloadIs(
						cc.From(user1Identity). // call by any user
									Query(balance.BalanceServiceChaincode_GetBalance,
								&balance.BalanceId{Address: ownerAddress, Symbol: erc20.Symbol}),
						&balance.Balance{}).(*balance.Balance)

					Expect(b.Address).To(Equal(ownerAddress))
					Expect(b.Amount).To(Equal(uint64(erc20.Token.TotalSupply)))
				})

				It(`Allow to get zero balance`, func() {
					b := expectcc.PayloadIs(
						cc.From(user1Identity).
							Query(balance.BalanceServiceChaincode_GetBalance,
								&balance.BalanceId{Address: user1Address, Symbol: erc20.Symbol}),
						&balance.Balance{}).(*balance.Balance)

					Expect(b.Amount).To(Equal(uint64(0)))
				})

			})

			var transferAmount uint64 = 100

			Context(`transfer`, func() {

				It(`Disallow to transfer balance by user with zero balance`, func() {
					expectcc.ResponseError(
						cc.From(user1Identity).
							Invoke(balance.BalanceServiceChaincode_Transfer,
								&balance.TransferRequest{
									Recipient: user2Address,
									Symbol:    erc20.Symbol,
									Amount:    transferAmount,
								}), balance.ErrAmountInsuficcient)

				})

				It(`Allow to transfer balance by owner`, func() {
					r := expectcc.PayloadIs(
						cc.From(ownerIdentity).
							Invoke(balance.BalanceServiceChaincode_Transfer,
								&balance.TransferRequest{
									Recipient: user1Address,
									Symbol:    erc20.Symbol,
									Amount:    transferAmount,
								}),
						&balance.TransferResponse{}).(*balance.TransferResponse)

					Expect(r.Sender).To(Equal(ownerAddress))
					Expect(r.Amount).To(Equal(transferAmount))
				})

				It(`Allow to get new non zero balance`, func() {
					b := expectcc.PayloadIs(
						cc.From(user1Identity).
							Query(balance.BalanceServiceChaincode_GetBalance,
								&balance.BalanceId{Address: user1Address, Symbol: erc20.Symbol}),
						&balance.Balance{}).(*balance.Balance)

					Expect(b.Amount).To(Equal(transferAmount))
				})

			})

			//todo: REFACTOR with new cc instance
			Context(`transfer batch `, func() {
				It(`Allow to transfer to 2 addresses`, func() {
					r := expectcc.PayloadIs(
						cc.From(ownerIdentity).
							Invoke(balance.BalanceServiceChaincode_TransferBatch,
								&balance.TransferBatchRequest{
									Transfers: []*balance.TransferRequest{{
										Recipient: user1Address,
										Symbol:    erc20.Symbol,
										Amount:    50,
									}, {
										Recipient: user2Address,
										Symbol:    erc20.Symbol,
										Amount:    150,
									}}}),
						&balance.TransferBatchResponse{}).(*balance.TransferBatchResponse)

					Expect(r.Transfers).To(HaveLen(2))
				})

				It(`Allow to get new non zero balance`, func() {
					b := expectcc.PayloadIs(
						cc.From(user1Identity).
							Query(balance.BalanceServiceChaincode_GetBalance,
								&balance.BalanceId{Address: user1Address, Symbol: erc20.Symbol}),
						&balance.Balance{}).(*balance.Balance)

					Expect(b.Amount).To(Equal(transferAmount + 50))

					b = expectcc.PayloadIs(
						cc.From(user2Identity).
							Query(balance.BalanceServiceChaincode_GetBalance,
								&balance.BalanceId{Address: user2Address, Symbol: erc20.Symbol}),
						&balance.Balance{}).(*balance.Balance)

					Expect(b.Amount).To(Equal(uint64(150)))
				})
			})

			Context(`Allowance`, func() {

				var allowAmount uint64 = 50

				It(`Allow to approve amount by owner for spender even if balance is zero`, func() {
					a := expectcc.PayloadIs(
						cc.From(user2Identity).
							Invoke(allowance.AllowanceServiceChaincode_Approve,
								&allowance.ApproveRequest{
									Owner:   user2Address,
									Spender: user1Address,
									Symbol:  erc20.Symbol,
									Amount:  allowAmount,
								}),
						&allowance.Allowance{}).(*allowance.Allowance)

					Expect(a.Owner).To(Equal(user2Address))
					Expect(a.Spender).To(Equal(user1Address))
					Expect(a.Amount).To(Equal(allowAmount))
				})
				It(`Disallow to approve amount by non owner`, func() {
					expectcc.ResponseError(
						cc.From(user2Identity).
							Invoke(allowance.AllowanceServiceChaincode_Approve,
								&allowance.ApproveRequest{
									Owner:   ownerAddress,
									Spender: user1Address,
									Symbol:  erc20.Symbol,
									Amount:  allowAmount,
								}), allowance.ErrOwnerOnly)
				})

				It(`Allow to approve amount by owner for spender if amount is sufficient`, func() {
					a := expectcc.PayloadIs(
						cc.From(ownerIdentity).
							Invoke(allowance.AllowanceServiceChaincode_Approve,
								&allowance.ApproveRequest{
									Owner:   ownerAddress,
									Spender: user2Address,
									Symbol:  erc20.Symbol,
									Amount:  allowAmount,
								}),
						&allowance.Allowance{}).(*allowance.Allowance)

					Expect(a.Owner).To(Equal(ownerAddress))
					Expect(a.Spender).To(Equal(user2Address))
					Expect(a.Amount).To(Equal(allowAmount))
				})

				It(`Allow to transfer from owner to spender`, func() {
					spenderIdentity := user2Identity
					spenderAddress := user2Address

					t := expectcc.PayloadIs(
						cc.From(spenderIdentity).
							Invoke(allowance.AllowanceServiceChaincode_TransferFrom,
								&allowance.TransferFromRequest{
									Owner:     ownerAddress,
									Recipient: spenderAddress,
									Symbol:    erc20.Symbol,
									Amount:    allowAmount,
								}),
						&allowance.TransferFromResponse{}).(*allowance.TransferFromResponse)

					Expect(t.Owner).To(Equal(ownerAddress))
					Expect(t.Recipient).To(Equal(spenderAddress))
					Expect(t.Amount).To(Equal(allowAmount))
				})
			})

			Context(`Burn`, func() {

				var burnAmount uint64 = 75

				It(`Allow to burn by owner`, func() {
					b := expectcc.PayloadIs(
						cc.From(user2Identity).
							Invoke(burnable.BurnableServiceChaincode_Burn,
								&burnable.BurnRequest{
									Address: user2Address,
									Symbol:  erc20.Symbol,
									Amount:  burnAmount,
								}),
						&burnable.BurnResponse{}).(*burnable.BurnResponse)

					Expect(b.Address).To(Equal(user2Address))
					Expect(b.Amount).To(Equal(burnAmount))

					balance := expectcc.PayloadIs(
						cc.From(user2Identity).
							Query(balance.BalanceServiceChaincode_GetBalance,
								&balance.BalanceId{Address: user2Address, Symbol: erc20.Symbol}),
						&balance.Balance{}).(*balance.Balance)

					Expect(balance.Amount).To(Equal(uint64(200 - burnAmount)))
				})

			})
		})
	}
}
