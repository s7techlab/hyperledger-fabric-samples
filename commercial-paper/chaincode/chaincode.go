package chaincode

import (
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/s7techlab/cckit/extensions/owner"
	"github.com/s7techlab/cckit/router"
	"github.com/s7techlab/cckit/state"
	m "github.com/s7techlab/cckit/state/mapping"
	"github.com/s7techlab/hyperledger-fabric-samples/commercial-paper/proto"
)

//  New create new chaincode struct with Init and Invoke methods
func New() (*router.Chaincode, error) {
	r := router.New(`commercial paper`)
	// Store on the ledger the information about chaincode instantiation
	r.Init(owner.InvokeSetFromCreator)

	if err := proto.RegisterCPaperChaincode(r, &CPaperImpl{}); err != nil {
		return nil, err
	}

	return router.NewChaincode(r), nil
}

type CPaperImpl struct {
}

// state wrapper with mappings defined
func (cc *CPaperImpl) state(ctx router.Context) m.MappedState {
	return m.WrapState(ctx.State(), m.StateMappings{}.
		//  Create mapping for Commercial Paper entity
		Add(&proto.CommercialPaper{},
			m.PKeySchema(&proto.CommercialPaperId{}), // Key namespace will be <"CommercialPaper", Issuer, PaperNumber>
			m.List(&proto.CommercialPaperList{}),     // Structure of result for List method
			m.UniqKey("ExternalId"),                  // External Id is unique
		))
}

// event wrapper with mappings defined
func (cc *CPaperImpl) event(ctx router.Context) state.Event {
	return m.WrapEvent(ctx.Event(), m.EventMappings{}.
		// Event name will be "IssueCommercialPaper", payload - same as issue payload
		Add(&proto.IssueCommercialPaper{}).
		// Event name will be "BuyCommercialPaper"
		Add(&proto.BuyCommercialPaper{}).
		// Event name will be "RedeemCommercialPaper"
		Add(&proto.RedeemCommercialPaper{}))
}

func (cc *CPaperImpl) List(ctx router.Context, in *empty.Empty) (*proto.CommercialPaperList, error) {
	// List method retrieves all entries from the ledger using GetStateByPartialCompositeKey method and passing it the
	// namespace of our contract type, in this example that's "CommercialPaper", then it unmarshals received bytes via
	// proto.Ummarshal method and creates a []proto.CommercialPaperList as defined in the
	// "StateMappings" variable at the top of the file
	if res, err := cc.state(ctx).List(&proto.CommercialPaper{}); err != nil {
		return nil, err
	} else {
		return res.(*proto.CommercialPaperList), nil
	}
}

func (cc *CPaperImpl) Get(ctx router.Context, id *proto.CommercialPaperId) (*proto.CommercialPaper, error) {
	if res, err := cc.state(ctx).Get(id, &proto.CommercialPaper{}); err != nil {
		return nil, err
	} else {
		return res.(*proto.CommercialPaper), nil
	}
}

func (cc *CPaperImpl) GetByExternalId(ctx router.Context, id *proto.ExternalId) (*proto.CommercialPaper, error) {
	if res, err := cc.state(ctx).GetByUniqKey(
		&proto.CommercialPaper{}, "ExternalId", []string{id.Id}, &proto.CommercialPaper{}); err != nil {
		return nil, err
	} else {
		return res.(*proto.CommercialPaper), nil
	}
}

func (cc *CPaperImpl) Issue(ctx router.Context, issue *proto.IssueCommercialPaper) (*proto.CommercialPaper, error) {
	// Validate input message using the rules defined in schema
	if err := issue.Validate(); err != nil {
		return nil, errors.Wrap(err, "payload validation")
	}

	// Create state entry
	cpaper := &proto.CommercialPaper{
		Issuer:       issue.Issuer,
		PaperNumber:  issue.PaperNumber,
		Owner:        issue.Issuer,
		IssueDate:    issue.IssueDate,
		MaturityDate: issue.MaturityDate,
		FaceValue:    issue.FaceValue,
		State:        proto.CommercialPaper_ISSUED, // Initial state
		ExternalId:   issue.ExternalId,
	}

	if err := cc.event(ctx).Set(issue); err != nil {
		return nil, err
	}

	if err := cc.state(ctx).Insert(cpaper); err != nil {
		return nil, err
	}
	return cpaper, nil
}

func (cc *CPaperImpl) Buy(ctx router.Context, buy *proto.BuyCommercialPaper) (*proto.CommercialPaper, error) {
	// Get the current commercial paper state
	cpaper, err := cc.Get(ctx, &proto.CommercialPaperId{Issuer: buy.Issuer, PaperNumber: buy.PaperNumber})
	if err != nil {
		return nil, errors.Wrap(err, "get cpaper")
	}

	// Validate current owner
	if cpaper.Owner != buy.CurrentOwner {
		return nil, fmt.Errorf(
			"paper %s %s is not owned by %s",
			cpaper.Issuer, cpaper.PaperNumber, buy.CurrentOwner)
	}

	// First buyData moves state from ISSUED to TRADING
	if cpaper.State == proto.CommercialPaper_ISSUED {
		cpaper.State = proto.CommercialPaper_TRADING
	}

	// Check paper is not already REDEEMED
	if cpaper.State == proto.CommercialPaper_TRADING {
		cpaper.Owner = buy.NewOwner
	} else {
		return nil, fmt.Errorf(
			"paper %s %s is not trading.current state = %s",
			cpaper.Issuer, cpaper.PaperNumber, cpaper.State)
	}

	if err = cc.event(ctx).Set(buy); err != nil {
		return nil, err
	}

	if err = cc.state(ctx).Put(cpaper); err != nil {
		return nil, err
	}

	return cpaper, nil
}

func (cc *CPaperImpl) Redeem(ctx router.Context, redeem *proto.RedeemCommercialPaper) (*proto.CommercialPaper, error) {
	// Get the current commercial paper state
	cpaper, err := cc.Get(ctx, &proto.CommercialPaperId{Issuer: redeem.Issuer, PaperNumber: redeem.PaperNumber})
	if err != nil {
		return nil, errors.Wrap(err, "get cpaper")
	}
	if err != nil {
		return nil, errors.Wrap(err, "paper not found")
	}

	// Check paper is not REDEEMED
	if cpaper.State == proto.CommercialPaper_REDEEMED {
		return nil, fmt.Errorf("paper %s %s is already redeemed", cpaper.Issuer, cpaper.PaperNumber)
	}

	// Verify that the redeemer owns the commercial paper before redeeming it
	if cpaper.Owner == redeem.RedeemingOwner {
		cpaper.Owner = redeem.Issuer
		cpaper.State = proto.CommercialPaper_REDEEMED
	} else {
		return nil, fmt.Errorf("redeeming owner does not own paper %s %s", cpaper.Issuer, cpaper.PaperNumber)
	}

	if err = cc.event(ctx).Set(redeem); err != nil {
		return nil, err
	}

	if err = cc.state(ctx).Put(cpaper); err != nil {
		return nil, err
	}

	return cpaper, nil
}

func (cc *CPaperImpl) Delete(ctx router.Context, id *proto.CommercialPaperId) (*proto.CommercialPaper, error) {
	cpaper, err := cc.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "get cpaper")
	}

	if err = cc.state(ctx).Delete(id); err != nil {
		return nil, err
	}

	return cpaper, nil
}
