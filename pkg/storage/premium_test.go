package storage

import (
	"testing"
	"time"
)

func TestCanTransfer(t *testing.T) {
	origin := &PostgresGuild{
		GuildID: 123,
		Premium: 0,
	}
	dest := &PostgresGuild{
		GuildID: 321,
		Premium: 0,
	}
	err := CanTransfer(origin, dest)
	if err == nil {
		t.Error("can't transfer from a free tier server")
	}
	origin.Premium = 1
	err = CanTransfer(origin, dest)
	if err == nil {
		t.Error("can't transfer a server with no transaction details")
	}
	var tt = int32(0)
	origin.TxTimeUnix = &tt
	err = CanTransfer(origin, dest)
	if err == nil {
		t.Error("can't transfer a server with expired premium")
	}
	tt = int32(time.Now().Unix())
	err = CanTransfer(origin, dest)

	// valid transfer
	if err != nil {
		t.Error(err)
	}

	origin.TransferredTo = &dest.GuildID
	err = CanTransfer(origin, dest)
	if err == nil {
		t.Error("can't transfer a server that has already been transferred")
	}

	origin.TransferredTo = nil
	origin.InheritsFrom = &dest.GuildID
	err = CanTransfer(origin, dest)
	if err == nil {
		t.Error("can't transfer a server that inherits status from another")
	}
	origin.InheritsFrom = nil
	dest.TransferredTo = &origin.GuildID
	err = CanTransfer(origin, dest)
	if err == nil {
		t.Error("can't transfer to a server that has transferred its status to another")
	}

	dest.TransferredTo = nil
	dest.InheritsFrom = &origin.GuildID
	err = CanTransfer(origin, dest)
	if err == nil {
		t.Error("can't transfer to a server that inherits its status from another")
	}

	dest.InheritsFrom = nil
	dest.Premium = 2
	err = CanTransfer(origin, dest)
	if err == nil {
		t.Error("can't transfer to a server that has existing non-standard premium")
	}

	dest.TxTimeUnix = &tt
	err = CanTransfer(origin, dest)
	if err == nil {
		t.Error("can't transfer to a server with active premium")
	}

	var ttt = int32(0)
	dest.TxTimeUnix = &ttt
	err = CanTransfer(origin, dest)
	if err != nil {
		t.Error(err)
	}
}
