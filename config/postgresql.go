package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func ConnectDB() (*pgx.Conn, context.Context, context.CancelFunc) {
	logger := GetLog()
	var username string = "postgres"
	var password string = "admin123"
	var host string = "103.212.39.79:31365"
	var database string = "pond-test"
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s", username, password, host, database)
	logger.Debug("Connection String : " + connStr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		cancel()
		logger.DPanic("Unable to connect to database : " + err.Error())
	}

	return conn, ctx, cancel
}
