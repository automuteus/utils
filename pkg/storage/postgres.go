package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/automuteus/utils/pkg/premium"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

type PsqlInterface struct {
	Pool *pgxpool.Pool

	// TODO does this require a lock? How should stuff be written/read from psql in an async way? Is this even a concern?
	//https://brandur.org/postgres-connections
}

func ConstructPsqlConnectURL(addr, username, password string) string {
	return fmt.Sprintf("postgres://%s?user=%s&password=%s", addr, username, password)
}

type PsqlParameters struct {
	Addr     string
	Username string
	Password string
}

func (psqlInterface *PsqlInterface) Init(addr string) error {
	dbpool, err := pgxpool.Connect(context.Background(), addr)
	if err != nil {
		return err
	}
	psqlInterface.Pool = dbpool
	return nil
}

func (psqlInterface *PsqlInterface) LoadAndExecFromFile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	tag, err := psqlInterface.Pool.Exec(context.Background(), string(bytes))
	if err != nil {
		return err
	}
	log.Println(tag.String())
	return nil
}

func (psqlInterface *PsqlInterface) insertGuild(guildID uint64, guildName string) error {
	_, err := psqlInterface.Pool.Exec(context.Background(), "INSERT INTO guilds VALUES ($1, $2, 0);", guildID, guildName)
	return err
}

func (psqlInterface *PsqlInterface) getGuild(guildID uint64) (*PostgresGuild, error) {
	var guilds []*PostgresGuild
	err := pgxscan.Select(context.Background(), psqlInterface.Pool, &guilds, "SELECT * FROM guilds WHERE guild_id=$1", guildID)
	if err != nil {
		return nil, err
	}

	if len(guilds) > 0 {
		return guilds[0], nil
	}
	return nil, errors.New("no guild found by that ID")
}

func (psqlInterface *PsqlInterface) insertUser(userID uint64) error {
	_, err := psqlInterface.Pool.Exec(context.Background(), "INSERT INTO users VALUES ($1, true);", userID)
	return err
}

func (psqlInterface *PsqlInterface) OptUserByString(userID string, opt bool) error {
	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return err
	}
	user, err := psqlInterface.EnsureUserExists(uid)
	if err != nil {
		return err
	}
	if user.Opt == opt {
		return errors.New("user opt status is already set to the value specified")
	}
	_, err = psqlInterface.Pool.Exec(context.Background(), "UPDATE users SET opt = $1 WHERE user_id = $2;", opt, uid)
	if err != nil {
		return err
	}
	if !opt {
		_, err = psqlInterface.Pool.Exec(context.Background(), "UPDATE game_events SET user_id = NULL WHERE user_id = $1;", uid)
		if err != nil {
			return err
		}

		_, err = psqlInterface.Pool.Exec(context.Background(), "DELETE FROM users_games WHERE user_id = $1;", uid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (psqlInterface *PsqlInterface) GetUserByString(userID string) (*PostgresUser, error) {
	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}
	return psqlInterface.GetUser(uid)
}

func (psqlInterface *PsqlInterface) GetUser(userID uint64) (*PostgresUser, error) {
	var users []*PostgresUser
	err := pgxscan.Select(context.Background(), psqlInterface.Pool, &users, "SELECT * FROM users WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	if len(users) > 0 {
		return users[0], nil
	}
	return nil, nil
}

func (psqlInterface *PsqlInterface) GetGame(guildID, connectCode, matchID string) (*PostgresGame, error) {
	var games []*PostgresGame
	err := pgxscan.Select(context.Background(), psqlInterface.Pool, &games, "SELECT * FROM games WHERE guild_id = $1 AND game_id = $2 AND connect_code = $3;", guildID, matchID, connectCode)
	if err != nil {
		return nil, err
	}
	if len(games) > 0 {
		return games[0], nil
	}
	return nil, nil
}

func (psqlInterface *PsqlInterface) GetGameEvents(matchID string) ([]*PostgresGameEvent, error) {
	var events []*PostgresGameEvent
	err := pgxscan.Select(context.Background(), psqlInterface.Pool, &events, "SELECT * FROM game_events WHERE game_id = $1 ORDER BY event_id ASC;", matchID)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (psqlInterface *PsqlInterface) insertGame(game *PostgresGame) (uint64, error) {
	t, err := psqlInterface.Pool.Query(context.Background(), "INSERT INTO games VALUES (DEFAULT, $1, $2, $3, $4, $5) RETURNING game_id;", game.GuildID, game.ConnectCode, game.StartTime, game.WinType, game.EndTime)
	if t != nil {
		for t.Next() {
			g := uint64(0)
			err := t.Scan(&g)

			if err != nil {
				log.Println(err)
				t.Close()
				return 0, err
			}
			t.Close()
			return g, nil
		}
	}
	return 0, err
}

func (psqlInterface *PsqlInterface) updateGame(gameID int64, winType int16, endTime int64) error {
	_, err := psqlInterface.Pool.Exec(context.Background(), "UPDATE games SET (win_type, end_time) = ($1, $2) WHERE game_id = $3;", winType, endTime, gameID)
	return err
}

func (psqlInterface *PsqlInterface) insertPlayer(player *PostgresUserGame) error {
	_, err := psqlInterface.Pool.Exec(context.Background(), "INSERT INTO users_games VALUES ($1, $2, $3, $4, $5, $6, $7);", player.UserID, player.GuildID, player.GameID, player.PlayerName, player.PlayerColor, player.PlayerRole, player.PlayerWon)
	return err
}

const SecsInADay = 86400

func (psqlInterface *PsqlInterface) GetGuildPremiumStatus(guildID string) (premium.Tier, int) {
	return psqlInterface.getGuildPremiumStatus(guildID, 0)
}

func (psqlInterface *PsqlInterface) getGuildPremiumStatus(guildID string, depth int) (premium.Tier, int) {
	// if we somehow recurse too deep...
	if depth > 3 {
		return premium.FreeTier, 0
	}
	// self-hosting; only return the true guild status if this variable is set
	if os.Getenv("AUTOMUTEUS_OFFICIAL") == "" {
		return premium.SelfHostTier, premium.NoExpiryCode
	}

	gid, err := strconv.ParseUint(guildID, 10, 64)
	if err != nil {
		log.Println(err)
		return premium.FreeTier, 0
	}

	guild, err := psqlInterface.getGuild(gid)
	if err != nil {
		return premium.FreeTier, 0
	}

	// if this is a recursive call, then we ignore the transfer (this is how inheriting works)
	if depth == 0 {
		// transferred servers are always treated as free tier, even if their tier/expiry is marked otherwise (the server
		// that premium was transferred to still uses these values, as "inherited")
		if guild.TransferredTo != nil {
			return premium.FreeTier, 0
		}
	}

	daysRem := premium.NoExpiryCode

	if guild.TxTimeUnix != nil {
		diff := time.Now().Unix() - int64(*guild.TxTimeUnix)
		// 31 - days elapsed
		daysRem = int(premium.SubDays - (diff / SecsInADay))
		// if the premium for this server is still active, return it (disregarding inheritance)
		if daysRem > 0 {
			return premium.Tier(guild.Premium), daysRem
		}
	}

	// follow the link to the inherited server
	// other tooling that facilitates transfers/gold sub-servers will need to be careful to avoid cyclic inheritance...
	if guild.InheritsFrom != nil {
		return psqlInterface.getGuildPremiumStatus(fmt.Sprintf("%d", *guild.InheritsFrom), depth+1)
	}

	return premium.Tier(guild.Premium), daysRem
}

func (psqlInterface *PsqlInterface) EnsureGuildExists(guildID uint64, guildName string) (*PostgresGuild, error) {
	guild, err := psqlInterface.getGuild(guildID)

	if guild == nil {
		err := psqlInterface.insertGuild(guildID, guildName)
		if err != nil {
			return nil, err
		}
		return psqlInterface.getGuild(guildID)
	}
	return guild, err
}

func (psqlInterface *PsqlInterface) EnsureUserExists(userID uint64) (*PostgresUser, error) {
	user, err := psqlInterface.GetUser(userID)

	if user == nil {
		err := psqlInterface.insertUser(userID)
		if err != nil {
			log.Println(err)
		}
		return psqlInterface.GetUser(userID)
	}
	return user, err
}

func (psqlInterface *PsqlInterface) AddInitialGame(game *PostgresGame) (uint64, error) {
	return psqlInterface.insertGame(game)
}

func (psqlInterface *PsqlInterface) AddEvent(event *PostgresGameEvent) error {
	if event.UserID == nil {
		_, err := psqlInterface.Pool.Exec(context.Background(), "INSERT INTO game_events VALUES (DEFAULT, NULL, $1, $2, $3, $4);", event.GameID, event.EventTime, event.EventType, event.Payload)
		return err
	}
	_, err := psqlInterface.Pool.Exec(context.Background(), "INSERT INTO game_events VALUES (DEFAULT, $1, $2, $3, $4, $5);", event.UserID, event.GameID, event.EventTime, event.EventType, event.Payload)
	return err
}

// make sure to call the relevant "ensure" methods before this one...
func (psqlInterface *PsqlInterface) UpdateGameAndPlayers(gameID int64, winType int16, endTime int64, players []*PostgresUserGame) error {
	err := psqlInterface.updateGame(gameID, winType, endTime)
	if err != nil {
		return err
	}

	for _, player := range players {
		err := psqlInterface.insertPlayer(player)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (psqlInterface *PsqlInterface) Close() {
	psqlInterface.Pool.Close()
}
