package postgresdb

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/patric-chuzhbe/gophkeeper/internal/models"
)

type PostgresDB struct {
	database *sql.DB
}

var ErrUserExists = errors.New("user already exists")

/*< new funcs>*/
func (db *PostgresDB) RegisterUser(ctx context.Context, login string, password string) error {
	_, err := db.database.ExecContext(ctx,
		`
			INSERT INTO users (user_id, login, pass)
				VALUES (gen_random_uuid(), $1, $2)
		`,
		login,
		password,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func (db *PostgresDB) GetUserRecords(ctx context.Context, userID string) (map[string]models.UserRecord, error) {
	rows, err := db.database.QueryContext(
		ctx,
		`
			SELECT 
				vault_records.record_id,
				vault_records.data, 
				vault_records.type, 
				vault_records.nonce,
				metadata.name,
				metadata.value
				FROM vault_records
					LEFT JOIN metadata ON 
						metadata.record_id = vault_records.record_id
				WHERE vault_records.user_id = $1
		`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]models.UserRecord{}
	for rows.Next() {
		var recordID string
		var typ string
		var encodedData string
		var nonce string
		var metadataName sql.NullString
		var metadataValue sql.NullString
		err := rows.Scan(
			&recordID,
			&encodedData,
			&typ,
			&nonce,
			&metadataName,
			&metadataValue,
		)
		if err != nil {
			return nil, err
		}
		record, ok := result[recordID]
		if !ok {
			decodedData, err := base64.StdEncoding.DecodeString(encodedData)
			if err != nil {
				return nil, err
			}
			decodedNonce, err := base64.StdEncoding.DecodeString(nonce)
			if err != nil {
				return nil, err
			}
			record = models.UserRecord{
				Type:  typ,
				Data:  decodedData,
				Nonce: decodedNonce,
				Meta:  map[string]string{},
			}
		}
		if metadataName.Valid {
			record.Meta[metadataName.String] = metadataValue.String
		}
		result[recordID] = record
	}
	return result, nil
}

func (db *PostgresDB) InsertRecord(
	ctx context.Context,
	userID string,
	data []byte,
	nonce []byte,
	recType string,
	meta map[string]string,
	tx *sql.Tx,
) error {
	var recordID string
	err := tx.QueryRowContext(ctx,
		`
			INSERT INTO vault_records (record_id, user_id, data, nonce, type)
				VALUES (gen_random_uuid(), $1, $2, $3, $4)
				RETURNING record_id
		`,
		userID,
		base64.StdEncoding.EncodeToString(data),
		base64.StdEncoding.EncodeToString(nonce),
		recType,
	).Scan(&recordID)
	if err != nil {
		return fmt.Errorf("insert value failed: %w", err)
	}

	for key, val := range meta {
		_, err := tx.ExecContext(
			ctx,
			`
				INSERT INTO metadata (metadata_id, record_id, name, value)
			 		VALUES (gen_random_uuid(), $1, $2, $3)
			 `,
			recordID,
			key,
			val,
		)
		if err != nil {
			return fmt.Errorf("insert metadata failed: %w", err)
		}
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
