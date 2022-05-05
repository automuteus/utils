package storage

import (
	"github.com/automuteus/utils/pkg/premium"
	"github.com/pashagolub/pgxmock"
	"testing"
	"time"
)

const (
	UserID            = "123123123123123123"
	UserIDInt  uint64 = 123123123123123123
	GuildID           = "234234234234234234"
	GuildIDInt uint64 = 234234234234234234
)

func TestIsUserPremium_nilTopGG(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mock.ExpectQuery("^SELECT (.+) FROM users WHERE user_id = (.+)$").
		WithArgs(UserIDInt).
		WillReturnRows(
			pgxmock.NewRows([]string{"user_id", "opt", "vote_time_unix"}).
				AddRow(UserIDInt, true, nil)) //return the vote time being now

	prem, err := isUserPremium(mock, nil, UserID)
	if err != nil {
		t.Error(err)
	}
	if prem {
		t.Error("user should not be premium; no vote time set, and top.gg client is nil")
	}
	// no expectations for a nil top gg
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestIsUserPremium(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	var now = int32(time.Now().Unix())

	mock.ExpectQuery("^SELECT (.+) FROM users WHERE user_id = (.+)$").
		WithArgs(UserIDInt).
		WillReturnRows(
			pgxmock.NewRows([]string{"user_id", "opt", "vote_time_unix"}).
				AddRow(UserIDInt, true, &now)) //return the vote time being now

	// now we execute our method
	prem, err := isUserPremium(mock, nil, UserID)
	if err != nil {
		t.Error(err)
	}
	if !prem {
		t.Error("expected premium status to be set")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestIsUserOrGuildPremium(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	var now = int32(time.Now().Unix())

	mock.ExpectQuery("^SELECT (.+) FROM guilds WHERE guild_id = (.+)$").
		WithArgs(GuildIDInt).
		WillReturnRows(
			pgxmock.NewRows([]string{"guild_id", "guild_name", "premium", "tx_time_unix", "transferred_to", "inherits_from"}).
				AddRow(GuildIDInt, "Some Name", int16(0), nil, nil, nil))

	mock.ExpectQuery("^SELECT (.+) FROM users WHERE user_id = (.+)$").
		WithArgs(UserIDInt).
		WillReturnRows(
			pgxmock.NewRows([]string{"user_id", "opt", "vote_time_unix"}).
				AddRow(UserIDInt, true, &now)) //return the vote time being now

	// now we execute our method
	tier, days, err := guildOrUserPremium(mock, nil, GuildID, UserID)
	if err != nil {
		t.Error(err)
	}
	if tier != premium.TrialTier {
		t.Error("expected premium status to be the trial tier")
	}
	if days != premium.NoExpiryCode {
		t.Error("expected a no expiry premium status")
	}
	if premium.IsExpired(tier, days) {
		t.Error("Trial tier with noexpiry should not evaluate to expired")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
