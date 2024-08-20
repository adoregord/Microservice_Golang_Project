package test

import (
	"database/sql"
	provider "microservice_orchestrator/internal/provider/db"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBConnection_Positive(t *testing.T) {
	db, err := provider.DBConnection()
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Cleanup
	db.Close()
}

func TestDBConnection_Negative(t *testing.T) {
	// Simulate an invalid connection string
	invalidConnString := "postgres://invalid:invalid@localhost:5432/order_db"
	db, err := sql.Open("pgx", invalidConnString)
	assert.NoError(t, err) // Error belum muncul di sini karena koneksi belum diuji.

	// Attempt to ping the database to force a real connection attempt.
	err = db.Ping()
	assert.Error(t, err) // Seharusnya gagal di sini karena string koneksi tidak valid.

	// Close the db connection if it's not nil
	if db != nil {
		db.Close()
	}
}
