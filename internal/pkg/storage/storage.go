package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type Storage struct {
	db *sql.DB
}

func initDB(db *sql.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS external_source_citations (
			hostname TEXT,
			repository TEXT,
			typeName TEXT,
			occurrences INTEGER,
			PRIMARY KEY (hostname, repository, typeName)
		);

		CREATE TABLE IF NOT EXISTS repositories (
			repository TEXT,
			owner TEXT,
			url TEXT,
			lang TEXT,
			PRIMARY KEY (repository)
		);
		
		CREATE TABLE IF NOT EXISTS counter (
		    id INTEGER,
			pullCnt INTEGER,
			issueCnt INTEGER,
		    PRIMARY KEY (id)         
		);

		INSERT INTO counter(id, pullCnt, issueCnt)
		VALUES(0, 0, 0)
		ON CONFLICT(id)
		DO NOTHING;

		CREATE TABLE IF NOT EXISTS aliasChecking (
			repository TEXT,
			url TEXT,
			typeName TEXT,
		    withAlias BOOLEAN,
		    isClosed BOOLEAN,
		    DurationInSec FLOAT64,
		    PRIMARY KEY (repository, url, typeName)     
		);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Error creating table:", err)
		return
	}
}

func New() *Storage {
	db, err := sql.Open("sqlite3", "./github_occurrences.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
		return nil
	}

	initDB(db)

	return &Storage{
		db: db,
	}
}

func (s *Storage) Close() {
	err := s.db.Close()
	if err != nil {
		log.Fatal("Error closing database:", err)
		return
	}
}

func (s *Storage) InsertNewSource(hostname, repository, typename string, occurrence int) {
	query := `
		INSERT INTO external_source_citations(hostname, repository, typeName, occurrences)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(hostname, repository, typeName)
			DO UPDATE SET occurrences = occurrences + ?
	`

	_, err := s.db.Exec(query, hostname, repository, typename, occurrence, occurrence)
	if err != nil {
		log.Fatal("Error inserting data:", err)
		return
	}
}

func (s *Storage) InsertNewRepo(repository, owner, url, lang string) {
	query := `
		INSERT INTO repositories(repository, owner, url, lang)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(repository)
			DO NOTHING;
	`

	_, err := s.db.Exec(query, repository, owner, url, lang)
	if err != nil {
		log.Fatal("Error inserting data:", err)
		return
	}
}

func (s *Storage) IncrementPull(delta int) {
	query := `
		UPDATE counter
		SET pullCnt = pullCnt + ?;
	`

	_, err := s.db.Exec(query, delta)
	if err != nil {
		log.Fatal("Error inserting data:", err)
		return
	}
}

func (s *Storage) IncrementIssue(delta int) {
	query := `
		UPDATE counter
		SET issueCnt = issueCnt + ?;
	`

	_, err := s.db.Exec(query, delta)
	if err != nil {
		log.Fatal("Error inserting data:", err)
		return
	}
}

func (s *Storage) InsertNewAliasEntry(repository, url, typename string, withAlias, isClosed bool, seconds float64) {
	query := `
		INSERT INTO aliasChecking(repository, url, typeName, withAlias, isClosed, DurationInSec)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(repository, url, typeName)
			DO NOTHING;
	`

	_, err := s.db.Exec(query, repository, url, typename, withAlias, isClosed, seconds)
	if err != nil {
		log.Fatal("Error inserting data:", err)
		return
	}
}
