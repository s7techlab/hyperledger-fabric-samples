package testdata

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	testcc "github.com/s7techlab/cckit/testing"

	"github.com/s7techlab/hyperledger-fabric-samples/samples/cpaper"
)

var (
	Id1 = &cpaper.CommercialPaperId{
		Issuer:      "SomeIssuer",
		PaperNumber: "0001",
	}

	ExternalId1 = &cpaper.ExternalId{
		Id: "EXT0001",
	}

	Issue1 = &cpaper.IssueCommercialPaper{
		Issuer:       Id1.Issuer,
		PaperNumber:  Id1.PaperNumber,
		IssueDate:    ptypes.TimestampNow(),
		MaturityDate: testcc.MustProtoTimestamp(time.Now().AddDate(0, 2, 0)),
		FaceValue:    100000,
		ExternalId:   ExternalId1.Id,
	}

	Buy1 = &cpaper.BuyCommercialPaper{
		Issuer:       Id1.Issuer,
		PaperNumber:  Id1.PaperNumber,
		CurrentOwner: Id1.Issuer,
		NewOwner:     "SomeBuyer",
		Price:        95000,
		PurchaseDate: ptypes.TimestampNow(),
	}

	Redeem1 = &cpaper.RedeemCommercialPaper{
		Issuer:         Id1.Issuer,
		PaperNumber:    Id1.PaperNumber,
		RedeemingOwner: Buy1.NewOwner,
		RedeemDate:     ptypes.TimestampNow(),
	}

	CpaperInState1 = &cpaper.CommercialPaper{
		Issuer:       Id1.Issuer,
		Owner:        Id1.Issuer,
		State:        cpaper.CommercialPaper_STATE_ISSUED,
		PaperNumber:  Id1.PaperNumber,
		FaceValue:    Issue1.FaceValue,
		IssueDate:    Issue1.IssueDate,
		MaturityDate: Issue1.MaturityDate,
		ExternalId:   Issue1.ExternalId,
	}
)
