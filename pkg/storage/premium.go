package storage

import (
	"context"
	"errors"
	"github.com/automuteus/utils/pkg/premium"
	"log"
	"strconv"
	"time"
)

// CanTransfer determines the set of possible transfers for server premium
// it does NOT allow for chained transfers! Aka if A -> B, then B cannot transfer to C (nor back to A)
func CanTransfer(origin, dest *PostgresGuild) error {
	if origin == nil || dest == nil {
		return errors.New("nil origin or dest server")
	}

	if origin.Premium == int16(premium.FreeTier) {
		return errors.New("origin server is free tier and cannot be transferred")
	}

	if origin.TransferredTo != nil {
		return errors.New("origin server has already been transferred to another server")
	}

	if origin.InheritsFrom != nil {
		return errors.New("origin server inherits premium from another server and cannot be transferred")
	}

	if dest.TransferredTo != nil {
		return errors.New("destination server has already transferred premium elsewhere")
	}

	if dest.InheritsFrom != nil {
		return errors.New("destination server inherits premium from another server and cannot be transferred")
	}

	if origin.TxTimeUnix == nil {
		return errors.New("origin server has no associated transaction and cannot be transferred")
	} else {
		diff := time.Now().Unix() - int64(*origin.TxTimeUnix)
		daysRem := int(premium.SubDays - (diff / SecsInADay))
		if premium.IsExpired(premium.Tier(origin.Premium), daysRem) {
			return errors.New("origin server has expired premium and cannot be transferred")
		}
	}

	if dest.TxTimeUnix != nil {
		diff := time.Now().Unix() - int64(*dest.TxTimeUnix)
		daysRem := int(premium.SubDays - (diff / SecsInADay))
		if !premium.IsExpired(premium.Tier(dest.Premium), daysRem) {
			return errors.New("destination server has active premium and cannot be overwritten")
		} else {
			// destination has premium, but it is expired
		}
	} else if dest.Premium != int16(premium.FreeTier) {
		return errors.New("cannot transfer to a server with existing non-standard premium")
	}

	return nil
}

func (psqlInterface *PsqlInterface) TransferPremium(origin, dest string) error {
	originID, err := strconv.ParseUint(origin, 10, 64)
	if err != nil {
		return err
	}
	destID, err := strconv.ParseUint(dest, 10, 64)
	if err != nil {
		return err
	}
	originGuild, err := psqlInterface.getGuild(originID)
	if err != nil {
		return err
	}
	destGuild, err := psqlInterface.getGuild(destID)
	if err != nil {
		return err
	}

	err = CanTransfer(originGuild, destGuild)
	if err != nil {
		return err
	}

	err = psqlInterface.setGuildInheritsFrom(dest, origin)
	if err != nil {
		return err
	}
	err = psqlInterface.setGuildTransferredTo(origin, dest)
	if err != nil {
		return err
	}
	return nil
}

func (psqlInterface *PsqlInterface) setGuildTransferredTo(guildID, transferTo string) error {
	_, err := psqlInterface.Pool.Exec(context.Background(), "UPDATE guilds SET transferred_to = $2 WHERE guild_id = $1;", guildID, transferTo)
	if err != nil {
		return err
	}
	log.Printf("Marked guild %s as transferred to: %s\n", guildID, transferTo)
	return nil
}

func (psqlInterface *PsqlInterface) setGuildInheritsFrom(guildID, inheritsFrom string) error {
	_, err := psqlInterface.Pool.Exec(context.Background(), "UPDATE guilds SET inherits_from = $2 WHERE guild_id = $1;", guildID, inheritsFrom)
	if err != nil {
		return err
	}
	log.Printf("Marked guild %s as inheriting from %s\n", guildID, inheritsFrom)
	return nil
}
