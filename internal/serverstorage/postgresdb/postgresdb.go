package postgresdb

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/patric-chuzhbe/gophkeeper/internal/models"
)

type PostgresDB struct {
	database *sql.DB
}

/*< new funcs>*/
func (db *PostgresDB) GetUserRecords(ctx context.Context, userID string) ([]models.UserRecord, error) {
	rows, err := db.database.QueryContext(
		ctx,
		`SELECT data, type, nonce FROM vault_records WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []models.UserRecord{}
	for rows.Next() {
		var typ string
		var encodedData string
		var nonce string
		if err := rows.Scan(&encodedData, &typ, &nonce); err != nil {
			return nil, err
		}
		decodedData, err := base64.StdEncoding.DecodeString(encodedData)
		if err != nil {
			return nil, err
		}
		decodedNonce, err := base64.StdEncoding.DecodeString(nonce)
		if err != nil {
			return nil, err
		}
		result = append(result, models.UserRecord{
			Type:  typ,
			Data:  decodedData,
			Nonce: decodedNonce,
		})
	}
	return result, nil
}

func (db *PostgresDB) InsertRecord(
	ctx context.Context,
	userID string,
	data []byte,
	nonce []byte,
	recType string,
	tx *sql.Tx,
) error {
	_, err := tx.ExecContext(ctx,
		`
			INSERT INTO vault_records (record_id, user_id, data, nonce, type)
				VALUES (gen_random_uuid(), $1, $2, $3, $4)
		`,
		userID,
		base64.StdEncoding.EncodeToString(data),
		base64.StdEncoding.EncodeToString(nonce),
		recType,
	)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}

func (db *PostgresDB) Authenticate(ctx context.Context, login string, password string) (string, error) {
	row := db.database.QueryRowContext(ctx,
		`SELECT user_id FROM users WHERE login = $1 AND pass = $2`,
		login,
		password,
	)
	var userID string
	err := row.Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("invalid login or password")
	}
	return userID, err
}

func (db *PostgresDB) resetDB(ctx context.Context) error {
	_, err := db.database.ExecContext(
		ctx,
		`
			DO $$
			DECLARE
				r RECORD;
			BEGIN
				FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
					EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
				END LOOP;
			END $$;
		`,
	)
	if err != nil {
		return fmt.Errorf("in internal/db/postgresdb/postgresdb.go/resetDB(): error while `db.database.ExecContext()` calling: %w", err)
	}
	return nil
}

func (db *PostgresDB) CommitTransaction(transaction *sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred while committing transaction: %v", r)
		}
	}()

	return transaction.Commit()
}

func (db *PostgresDB) RollbackTransaction(transaction *sql.Tx) error {
	return transaction.Rollback()
}

func (db *PostgresDB) BeginTransaction() (*sql.Tx, error) {
	return db.database.Begin()
}

type InitOption func(*initOptions)

type initOptions struct {
	DBPreReset bool
}

func WithDBPreReset(value bool) InitOption {
	return func(options *initOptions) {
		options.DBPreReset = value
	}
}

func New(databaseDSN string, migrationsDir string, optionsProto ...InitOption) (*PostgresDB, error) {
	options := &initOptions{
		DBPreReset: false,
	}
	for _, protoOption := range optionsProto {
		protoOption(options)
	}

	database, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("in internal/db/postgresdb/postgresdb.go/New(): error while `sql.Open()` calling: %w", err)
	}

	result := &PostgresDB{
		database: database,
	}

	if options.DBPreReset {
		if err := result.resetDB(context.TODO()); err != nil {
			return nil, fmt.Errorf("in internal/db/postgresdb/postgresdb.go/New(): error while `result.resetDB()` calling: %w", err)
		}
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("in internal/db/postgresdb/postgresdb.go/New(): error while `goose.SetDialect()` calling: %w", err)
	}

	if err := goose.Up(result.database, migrationsDir); err != nil {
		return nil, fmt.Errorf("in internal/db/postgresdb/postgresdb.go/New(): error while `goose.Up()` calling: %w", err)
	}

	return result, nil
}

func (db *PostgresDB) Close() error {
	err := db.database.Close()
	if err != nil {
		return fmt.Errorf("in internal/db/postgresdb/postgresdb.go/Close(): error while `db.database.Close()` calling: %w", err)
	}

	return nil
}

/*</new funcs>*/
