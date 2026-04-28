package storage

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

type Reading struct {
	ID         int64
	SensorMAC  string
	SensorName string
	Type       string
	Value      float64
	Unit       string
	Timestamp  time.Time
}

func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS readings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sensor_mac TEXT NOT NULL,
			sensor_name TEXT NOT NULL,
			type TEXT NOT NULL,
			value REAL NOT NULL,
			unit TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_readings_timestamp ON readings(timestamp);
	`)
	return err
}

func (s *Storage) Save(reading *Reading) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO readings (sensor_mac, sensor_name, type, value, unit, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, reading.SensorMAC, reading.SensorName, reading.Type, reading.Value, reading.Unit, reading.Timestamp)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}
