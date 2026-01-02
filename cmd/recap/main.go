package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"garrison-stauffer.com/discord-bot/discord/api"
	"garrison-stauffer.com/discord-bot/environment"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
created db schema in the form:

create table if not exists
discord_messages_json (
    id              bigint      primary key,
    channel_id     bigint      not null,
    author_id      bigint      not null,
    content        text        not null,
    timestamp      timestamptz not null,
    raw_json       jsonb       not null
);

create table if not exists discord_channels_json (
    id          bigint      primary key,
    guild_id    bigint      null,
    name        text        not null,
    raw_json    jsonb       not null
);

create table if not exists discord_guilds_json (
	id          bigint      primary key,
	name        text        not null,
	raw_json    jsonb       not null
);

create table if not exists discord_users_json
(
    id          bigint primary key,
    nick        text  not null,
    username    text  not null,
    global_name text  not null,
    raw_json    jsonb not null
);

create table if not exists discord_message_reactions (
		message_id  bigint      not null,
		channel_id  bigint      not null,
		emoji_id    text        not null,
		emoji_name text         not null,
		count       integer     not null,
		raw_json  jsonb         not null,
		primary key (message_id, emoji_id, emoji_name)
)
*/

// I want to pull from this guild: 1023753903994572820

const (
	mumbleGuild = "1023753903994572820"
)

func main() {
	// getAllChannels()
	// getAllUsers()
	// scanMessages("1023753904455950409")
	// for _, channelId := range []string{
	// "1023754647841800253",
	// "1023758435591925860",
	// "1024307707655770193",
	// "1024317860971032636",
	// "1034560362697195580",
	// "1056472933314334800",
	// "1088809932041756763", #pax-2026 (private channel)
	// "1127623738918182922",
	// "1170072330605711370",
	// "1198948721782693960",
	// "1213655674152816710",
	// "1284719556148990054",
	// "1317943688126791751",
	// "1348738728318865419",
	// "1380272545609285802",
	// } {
	// 	scanMessages(channelId)
	// 	// sleep between channels to avoid rate limits
	// 	time.Sleep(2 * time.Second)
	// }

	fetchMessageReactions()
}

// fetchMessageReactions fetches reactions for messages in 2025 and stores them in the db.
// it queries the database to find message ids that have reactions
func fetchMessageReactions() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, "postgresql://garrison:@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}

	botSecret := environment.BotSecret()

	// this needs to unroll reactions and get the emoji for each one
	// it looks like this in the db: `[{"me": true, "count": 4, "emoji": {"id": null, "name": "❤️"}, "burst_me": false, "me_burst": false, "burst_count": 0, "burst_colors": [], "count_details": {"burst": 0, "normal": 4}}]`
	rows, err := db.Query(ctx, `
		with raw_reactions as (select id, channel_id, jsonb_array_elements(messages.raw_json -> 'reactions') as reaction
                       	from discord_messages_json messages),
		reactions as (select id, channel_id, (reaction->'emoji'->>'id') as emoji_id, (reaction->'emoji'->>'name') as emoji_name
						from raw_reactions)
		select id, channel_id, coalesce( emoji_name || ':' || emoji_id, emoji_name), coalesce(emoji_id, ''), coalesce(emoji_name, '') from reactions;
	`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	tx, err := db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		panic(err)
	}

	var total int = 0
	// flush reactions on some cadenece to avoid long running transactions
	for rows.Next() {
		// sleep a bit between requests to avoid rate limits
		// we should be able to do 5rps
		time.Sleep(1 * time.Second)

		var messageId string
		var channelId string
		var reactId string
		var emojiId string
		var emojiName string
		err := rows.Scan(&messageId, &channelId, &reactId, &emojiId, &emojiName)
		if err != nil {
			slog.Error("error scanning row", "error", err)
			continue
		}

		// slog.Info("fetching reactions for message", "message_id", messageId, "channel_id", channelId)
		req, err := api.NewListMessageReactions(channelId, messageId, url.QueryEscape(reactId), botSecret)
		if err != nil {
			slog.Error("error creating list message reactions request", "error", err)
			continue
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("error making list message reactions request", "error", err)
			continue
		}

		if res.StatusCode == http.StatusTooManyRequests {
			/*

				< HTTP/1.1 429 TOO MANY REQUESTS
				< Content-Type: application/json
				< Retry-After: 65
				< X-RateLimit-Limit: 10
				< X-RateLimit-Remaining: 0
				< X-RateLimit-Reset: 1470173023.123
				< X-RateLimit-Reset-After: 64.57
				< X-RateLimit-Bucket: abcd1234
				< X-RateLimit-Scope: user
				{
				"message": "You are being rate limited.",
				"retry_after": 64.57,
				"global": false
				}
			*/
			// I should use retry after or reset after to sleep for that many seconds before trying again
			slog.Warn("rate limited when fetching message reactions", "message_id", messageId, "channel_id", channelId, "headers", res.Header)
			res.Body.Close()
			retryAfter := res.Header.Get("Retry-After")
			retryAfterDur, err := time.ParseDuration(retryAfter + "s")
			if err != nil {
				slog.Error("error parsing retry after duration", "error", err)
				// default to 60 seconds
				retryAfterDur = 60 * time.Second
			}
			slog.Info("sleeping for retry after duration", "duration", retryAfterDur, "actually doing", 5*time.Second)
			time.Sleep(5 * time.Second)
			continue
		} else if res.StatusCode != http.StatusOK {
			slog.Error("non-200 response from list message reactions", "status_code", res.StatusCode, "message_id", messageId, "channel_id", channelId, "emoji_id", emojiId, "emoji_name", emojiName)
			res.Body.Close()
			continue
		}

		var reactions []map[string]any
		err = json.NewDecoder(res.Body).Decode(&reactions)
		res.Body.Close()

		// slog.Info("response header", "headers", res.Header)

		if err != nil {
			slog.Error("error decoding list message reactions response", "error", err)
			continue
		}

		// slog.Info("fetched reactions", "message_id", messageId, "reaction_count", len(reactions))
		// store reactions in the db
		// as the whole model, we can unnest later
		count := len(reactions)
		insert, err := tx.Exec(ctx, `
		insert into discord_message_reactions (message_id, channel_id, emoji_id, emoji_name, count, raw_json)
		values ($1, $2, $3, $4, $5, $6)
		on conflict (message_id, emoji_id, emoji_name) do update set
			count = excluded.count,
			raw_json = excluded.raw_json
		`, messageId, channelId, emojiId, emojiName, count, reactions)

		if err != nil {
			slog.Error("error inserting message reaction into db", "error", err, "message_id", messageId, "channel_id", channelId, "emoji_id", emojiId, "emoji_name", emojiName)
			continue
		}

		if insert.RowsAffected() == 0 {
			slog.Error("stored message reaction", "message_id", messageId, "channel_id", channelId, "emoji_id", emojiId, "emoji_name", emojiName, "count", count)
		}

		total++
		if total%50 == 0 {
			slog.Info("total reactions processed", "count", total)
			err := tx.Commit(ctx)
			if err != nil {
				slog.Error("error committing transaction", "error", err)
			}
			tx, err = db.BeginTx(ctx, pgx.TxOptions{
				IsoLevel: pgx.ReadCommitted,
			})
			if err != nil {
				panic(err)
			}
		}
	}
}

// scanMessages fetches all messages in a channel for 2025 and stores them in the db.
// conflicts in the db based on primary id should be ignored
func scanMessages(channelId string) {
	ctx := context.Background()
	db, err := pgx.Connect(ctx, "postgresql://garrison:@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close(ctx)

	botSecret := environment.BotSecret()

	// Start date for 2025
	startOf2025 := "2025-01-01T00:00:00.000000+00:00"

	var lastMessageId string
	totalMessages := 0

	for {
		// Build the request with pagination
		var req *http.Request
		if lastMessageId == "" {
			req, err = api.NewListMessages(channelId, botSecret, func(r *http.Request) {
				r.URL.RawQuery = "limit=100"
			})
		} else {
			req, err = api.NewListMessages(channelId, botSecret, func(r *http.Request) {
				r.URL.RawQuery = fmt.Sprintf("before=%s&limit=100", lastMessageId)
			})
		}

		if err != nil {
			slog.Error("error creating list messages request", "error", err)
			panic(err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("error making list messages request", "error", err)
			panic(err)
		}

		if res.StatusCode != http.StatusOK {
			slog.Error("non-200 response from list messages", "status_code", res.StatusCode)
			panic(fmt.Sprintf("non-200 response: %d", res.StatusCode))
		}

		var messages []map[string]any
		err = json.NewDecoder(res.Body).Decode(&messages)
		res.Body.Close()

		if err != nil {
			slog.Error("error decoding list messages response", "error", err)
			panic(err)
		}

		if len(messages) == 0 {
			slog.Info("no more messages to fetch")
			break
		}

		// Process each message
		for _, message := range messages {
			timestamp := message["timestamp"].(string)

			// Check if we've reached messages before 2025
			if timestamp < startOf2025 {
				slog.Info("reached messages before 2025, stopping", "timestamp", timestamp)
				slog.Info("total messages scanned", "count", totalMessages)
				return
			}

			author := message["author"].(map[string]any)

			_, err := db.Exec(ctx, `
				insert into discord_messages_json (id, channel_id, author_id, content, timestamp, raw_json)
				values ($1, $2, $3, $4, $5, $6)
				on conflict (id) do update set
					channel_id = excluded.channel_id,
					author_id = excluded.author_id,
					content = excluded.content,
					timestamp = excluded.timestamp,
					raw_json = excluded.raw_json
			`, message["id"], message["channel_id"], author["id"], message["content"], timestamp, message)

			if err != nil {
				slog.Error("error inserting message into db", "error", err, "message_id", message["id"])
			} else {
				totalMessages++
				if totalMessages%100 == 0 {
					slog.Info("progress", "messages_scanned", totalMessages)
				}
			}

			lastMessageId = message["id"].(string)
		}

		// also log the earliest message that we've found in this batch
		earliestMessage := messages[len(messages)-1]
		slog.Info("earliest message in batch", "id", earliestMessage["id"], "timestamp", earliestMessage["timestamp"])
		slog.Info("fetched batch", "count", len(messages), "total", totalMessages)
		// sleep for a short duration to avoid rate limits
		time.Sleep(1500 * time.Millisecond)
	}

	slog.Info("scan complete", "total_messages", totalMessages)
}

func getAllUsers() {
	ctx := context.Background()
	foo, err := pgx.Connect(ctx, "postgresql://garrison:@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer foo.Close(ctx)

	botSecret := environment.BotSecret()
	members, err := api.NewListGuildMembers(mumbleGuild, botSecret, func(req *http.Request) {
		req.URL.RawQuery = "limit=1000"
	})
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(members)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// process the response

	var response []map[string]any
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		slog.Error("error decoding list guild members response", "error", err)
		panic(err)
	}

	for _, member := range response {
		slog.Info("member", "member", member)

		user := member["user"].(map[string]any)

		nick := ""
		globalName := ""
		maybeNick := member["nick"]
		if maybeNick != nil {
			nick = maybeNick.(string)
		}
		maybeGlobalName := user["global_name"]
		if maybeGlobalName != nil {
			globalName = maybeGlobalName.(string)
		}

		_, err := foo.Exec(ctx, `
			insert into discord_users_json (id, nick, username, global_name, raw_json)
			values ($1, $2, $3, $4, $5)
			on conflict (id) do update set
				nick = excluded.nick,
				username = excluded.username,
				global_name = excluded.global_name,
				raw_json = excluded.raw_json
		`, user["id"], nick, user["username"], globalName, user)
		if err != nil {
			slog.Error("error inserting user into db", "error", err)
		}
	}
}

func getAllChannels() {
	ctx := context.Background()
	foo, err := pgx.Connect(ctx, "postgresql://garrison:@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer foo.Close(ctx)

	botSecret := environment.BotSecret()
	members, err := api.NewListGuildChannels(mumbleGuild, botSecret)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(members)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// process the response

	var response []map[string]any
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		slog.Error("error decoding list guild channels response", "error", err)
		panic(err)
	}

	for _, channel := range response {
		slog.Info("channel", "channel", channel)

		_, err := foo.Exec(ctx, `
			insert into discord_channels_json (id, guild_id, name, raw_json)
			values ($1, $2, $3, $4)
			on conflict (id) do update set
				guild_id = excluded.guild_id,
				name = excluded.name,
				raw_json = excluded.raw_json
		`, channel["id"], channel["guild_id"], channel["name"], channel)
		if err != nil {
			slog.Error("error inserting channel into db", "error", err)
		}
	}
}
