package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// The migration used to create the announcements table.
const spoolSchema = `
	CREATE TABLE IF NOT EXISTS announcements(
		id INTEGER PRIMARY KEY,
		channel_key VARCHAR(64) NOT NULL,
		video_id VARCHAR(64) NOT NULL,
		announced_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS announcements_videos ON announcements(channel_key, video_id);
`

// The command used to register a new announcement into the table.
const spoolInsert = `INSERT INTO announcements(channel_key, video_id) VALUES(?, ?)`

// The command used to query an announcement from the database.
const spoolQuery = `SELECT id FROM announcements WHERE channel_key = ? AND video_id = ?`

// The Spool holds information about which videos have been previously
// announced.  The spool exists to prevent a video from being notified twice,
// thus it should be checked that the video is in the spool before announcing
// it, videos should be added to the spool once they are announced
// successfully.
type Spool struct {
	db *sql.DB // The abcking database
}

// NewSpool opens a new spool file with the given path. Because the underlying
// engine is an SQLite database, `:memory:` is an acceptable value that will
// return an in-memory spool file. This is useful for testing, mostly.
func NewSpool(path string) (*Spool, error) {
	db, err := openSpool(path)
	if err != nil {
		return nil, err
	}
	return &Spool{db: db}, nil
}

// Close will close the spool file.
func (spool *Spool) Close() error {
	return spool.db.Close()
}

// MarkAsAnnounced will add the given video into the channel spool. Because
// a specific feed may have multiple references, the channel reference has
// to be given as parameter, so that multiple webhooks can broadcast the same
// video ID if required.
func (spool *Spool) MarkAsAnnounced(channel, video string) error {
	log.Printf("Marking %s as announced on channel %s", video, channel)

	stmt, err := spool.db.Prepare(spoolInsert)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(channel, video)
	return err
}

// IsAnnounced will check whether the given video has been announced in a
// specific channel spool. If MarkAsAnnounced was previously called with the
// same parameters, this function will return true.
func (spool *Spool) IsAnnounced(channel, video string) (bool, error) {
	stmt, err := spool.db.Prepare(spoolQuery)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	res, err := stmt.Query(channel, video)
	if err != nil {
		return false, err
	}
	defer res.Close()

	return res.Next(), nil // if there is a first row, then this video has been announced.
}

func openSpool(path string) (*sql.DB, error) {
	log.Printf("Opening spool file: %s", path)

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := initSpool(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func initSpool(db *sql.DB) error {
	log.Printf("Initializing spool schema...")

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(spoolSchema)
	if err != nil {
		return err
	}
	return tx.Commit()
}
