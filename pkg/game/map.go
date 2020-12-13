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
