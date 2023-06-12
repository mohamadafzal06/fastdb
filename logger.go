package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
)

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error
	ReadEvents() (<-chan Event, <-chan error)
	Run()
}

type FileTransactionLogger struct {
	// Write-only channel for sending events
	events chan<- Event
	// Read-only channel for receiving errors
	errors <-chan error
	// The last used event sequence number
	lastSequence uint64
	file         *os.File
}

func NewFileTransactionLogger(filename string) (*FileTransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w\n", err)
	}

	return &FileTransactionLogger{file: file}, nil
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *FileTransactionLogger) WritePut(key string, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *FileTransactionLogger) Run() {
	events := make(chan Event, 16)
	errors := make(chan error, 1)
	l.events = events
	l.errors = errors

	go func() {
		for e := range events {
			l.lastSequence++

			_, err := fmt.Fprintf(l.file,
				"%d\t%d\t%s\t%s\n",
				l.lastSequence, e.EventType, e.Key, e.Value)

			if err != nil {
				errors <- err
				return
			}
		}

	}()
}

func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file)
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		var e Event
		defer close(outEvent)
		defer close(outError)

		for scanner.Scan() {
			line := scanner.Text()

			if err := fmt.Errorf(line, "%d\t%d\t%s\t%s\n",
				&e.Sequence, &e.EventType, &e.Key, &e.Value); err != nil {
				outError <- fmt.Errorf("input parse error: %w", err)
				return
			}

			if l.lastSequence > e.Sequence {
				outError <- fmt.Errorf("transaction number out of sequence")
				return
			}

			l.lastSequence = e.Sequence

			outEvent <- e
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()

	return outEvent, outError
}

type PostgresTransactionLogger struct {
	events chan<- Event
	errors <-chan error
	db     *sql.DB
}

// TODO: add to config file
type PostgresDBConfig struct {
	host     string
	dbName   string
	user     string
	password string
}

func NewPostgresTransactionLogger(cf PostgresDBConfig) (*PostgresTransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s", cf.host, cf.dbName, cf.user, cf.password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("connect to db failed: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to open db connction: %w", err)
	}

	logger := &PostgresTransactionLogger{db: db}

	// TODO: Add this 2 methods
	//exists, err := logger.verifyTableExists()
	//if err != nil {
	//	return nil, fmt.Errorf("failed to verify table exists: %w", err)
	//}

	//if !exists {
	//	if err = <-logger.createTable(); err != nil {
	//		return nil, fmt.Errorf("failed to create table: %w", err)
	//	}
	//}

	return logger, nil

}

func (p *PostgresTransactionLogger) WriteDelete(key string) {
	p.events <- Event{EventType: EventDelete, Key: key}
}

func (p *PostgresTransactionLogger) WritePut(key string, value string) {
	p.events <- Event{EventType: EventDelete, Key: key, Value: value}
}

func (p *PostgresTransactionLogger) Err() <-chan error {
	return p.errors
}

func (p *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		var e Event
		defer close(outEvent)
		defer close(outError)

		query := `SELECT sequence, event_type, key, value FROM transaction ORDER BY sequence`

		// TODO: add context
		rows, err := p.db.Query(query)
		if err != nil {
			outError <- fmt.Errorf("sql query error: %w", err)
			return
		}

		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				outError <- fmt.Errorf("error reading row: %w", err)
				return
			}
			outEvent <- e
		}

		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()

	return outEvent, outError
}

func (p *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16)
	p.events = events

	errors := make(chan error, 1)
	p.errors = errors

	go func() {
		query := `INSER INTO transactions (evnet_type, key, value) VALUES ($1, $2, $3)`

		for e := range events {
			_, err := p.db.Exec(query, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
			}
		}
	}()
}
