package game

type PlayMap int

const (
	SKELD PlayMap = iota
	MIRA
	POLUS
	EMPTYMAP PlayMap = 10
)

var MapNames = map[PlayMap]string{
	SKELD: "Skeld",
	MIRA:  "Mira",
	POLUS: "Polus",
}

var NameToPlayMap = map[string]int32{
	"the_skeld": (int32)(SKELD),
	"mira_hq":   (int32)(MIRA),
	"polus":     (int32)(POLUS),
	"NoMap":     -1,
}
