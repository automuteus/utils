package task

type PremiumTier int16

const (
	FreeTier PremiumTier = iota
	BronzeTier
	SilverTier
	GoldTier
	PlatTier
	SelfHostTier
)

var PremiumTierStrings = []string{
	"Free",
	"Bronze",
	"Silver",
	"Gold",
	"Platinum",
	"SelfHost",
}
