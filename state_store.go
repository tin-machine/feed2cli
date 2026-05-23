package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

func newHatenaStateStore(options hatenaOutputOptions) (hatenaStateStore, error) {
	backend := strings.ToLower(strings.TrimSpace(options.StateBackend))
	if backend == "" {
		backend = envDefault("FEED2CLI_STATE_BACKEND", "json")
	}

	path := options.StatePath
	if path == "" {
		path = envDefault("FEED2CLI_STATE_PATH", "")
	}
	if path == "" {
		if backend == "sqlite" {
			path = "hatena_state.sqlite"
		} else {
			path = stateFilePath
		}
	}

	switch backend {
	case "json", "file":
		return fileHatenaStateStore{path: path}, nil
	case "sqlite", "sqlite3":
		return sqliteHatenaStateStore{path: path}, nil
	default:
		return nil, fmt.Errorf("unsupported state backend %q", backend)
	}
}

type sqliteHatenaStateStore struct {
	path string
}

func (store sqliteHatenaStateStore) Load() (State, error) {
	db, err := sql.Open("sqlite", store.path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if err := ensureHatenaStateSchema(db); err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT entry_url, last_comment_timestamp, slack_thread_timestamp
		FROM hatena_entry_state
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	state := make(State)
	for rows.Next() {
		var key string
		var value HatenaEntryState
		if err := rows.Scan(&key, &value.LastCommentTimestamp, &value.SlackThreadTimestamp); err != nil {
			return nil, err
		}
		state[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return state, nil
}

func (store sqliteHatenaStateStore) Save(state State) error {
	db, err := sql.Open("sqlite", store.path)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := ensureHatenaStateSchema(db); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM hatena_entry_state`); err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO hatena_entry_state (
			entry_url,
			last_comment_timestamp,
			slack_thread_timestamp
		) VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, value := range state {
		if _, err := stmt.Exec(key, value.LastCommentTimestamp, value.SlackThreadTimestamp); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func ensureHatenaStateSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS hatena_entry_state (
			entry_url TEXT PRIMARY KEY,
			last_comment_timestamp TEXT NOT NULL,
			slack_thread_timestamp TEXT NOT NULL
		)
	`)
	return err
}
