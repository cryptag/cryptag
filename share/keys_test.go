// Steve Phillips / elimisteve
// 2017.01.16

package share

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKeyPair(t *testing.T) {
	type pair struct {
		email string
		pass  string
		mID   string
	}

	tests := []pair{
		{
			"cc8942bc1684ed33f61e66dd89c9232b94e6518d6ff2060579753772ca41a3e86f5a083f91f78afc43820ebe28676a52@cryptag.org",
			"AngelfishEbookKabobAortaPurseFettuccineWaspIphoneSwimwearSproutNirvanaMolecule",

			"ZuhYoKx565EzEsn26FGguhSMu5oWuCivkiiPRCSnyHnbd",
		},
		{
			"b8874cca5c2ee70370e4bc32cfb54fab5d47e08fe8f70960e285fc9f2df58531140fbef092cdedc7be8056b104f57dfe@cryptag.org",
			"test passphrase",
			"i7U7qSKvEyYNdXHtCSighRRb2NWZ6cLzwNuriz8HffdxN",
		},
		{
			"abc6ea763dc214df8ea2e645ee045b7e439ffa28bb9ffb133c12a4945aceb048da86a8f3f7f79b607d358572112d1e43@cryptag.org",
			"some other passphrase PageantCajolingSweatshirtBoroughTulipAbandonedHazelnutTuxedoAvenue",
			"uRDKbbkCUKuMZ928zEf5Y8eaCBcbQTrCoum4LACkNGe2b",
		},
		{
			"af44b588789c3588116d54f827a48f2287554ee8096216a95e61edf4a5f44abad542b213f960738003dada4d87d6699f@cryptag.org",
			"TuskKioskSamuraiAtomBobcatZeppelinExplodeDesktopDegreePsychopathFondueSnugness",
			"5JjbLEaRJ2GFMRKGHsJ2cwxzKnL2wHsMB5Umk27Me1a4f",
		},
		{
			"e0ce2970d5f854a4a245c1ec16a5b556deb63ad54a7d7b879a884ca3479ea3c16d120e1fc5a7ddb2c56312c46da57891@cryptag.org",
			"RefineryDinnerwareAbhorrencePatioPyramidFleshinessIrrigationExcerptAlsoUsherHeadbandSmokestack",
			"zDpNXffd6X7fJ5xDbRuf3Ad9SmQQhnXcAye42moBLwBei",
		},
	}

	for _, tt := range tests {
		gotem := EmailFromPassphrase(tt.pass)
		assert.Equal(t, tt.email, gotem)

		keypair, _ := NewKeyPair(tt.pass)
		got, _ := keypair.EncodeID()
		assert.Equal(t, tt.mID, got)
	}
}
