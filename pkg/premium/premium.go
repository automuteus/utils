package premium

type Tier int16

const (
	FreeTier Tier = iota
	BronzeTier
	SilverTier
	GoldTier
	PlatTier
	SelfHostTier
)

var TierStrings = []string{
	"Free",
	"Bronze",
	"Silver",
	"Gold",
	"Platinum",
	"SelfHost",
}
