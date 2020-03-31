package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

type PostgresDB struct {
	pg *pg.DB
}

var _ Database = (*PostgresDB)(nil)

// TODO: [LATER] use generic options, not ones specific to pg package
func ConnectPostgres(options *pg.Options) (*PostgresDB, error) {
	db := &PostgresDB{
		pg: pg.Connect(options),
	}
	// TODO: enable SQL logger only if requested via flag
	db.pg.AddQueryHook(pgLogger{})

	// FIXME: how to add indexes via pg package?
	err := db.pg.CreateTable((*User)(nil), &orm.CreateTableOptions{
		// TODO: [LATER] switch to proper migrations, e.g. https://github.com/go-pg/migrations or https://github.com/robinjoseph08/go-pg-migrations
		IfNotExists: true,
	})
	if err != nil {
		db.pg.Close()
		return nil, fmt.Errorf("creating schemas: %w", err)
	}
	return db, nil
}

func (db *PostgresDB) Close() error {
	return db.pg.Close()
}

func (db *PostgresDB) ListUsers() ([]*User, error) {
	var users []*User
	err := db.pg.Model(&users).Select()
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	return users, nil
}

func (db *PostgresDB) GetUser(email string) (*User, error) {
	// TODO: [LATER] is there a smarter way to return 0..1 records with pg package?
	var users []*User
	err := db.pg.Model(&users).
		Where(`email = ?`, email).
		Where(`deleted IS NULL`).
		Select()
	if err != nil {
		log.Printf("GetUser: %#v", err)
		return nil, fmt.Errorf("getting user: %w", err)
	}

	switch len(users) {
	case 0:
		return nil, nil
	case 1:
		return users[0], nil
	default:
		// TODO: [LATER] emit an error ID to logs and to Errorf, for cross-referencing
		// TODO: [LATER] consider printing a few of the returned users to logs for easier debugging (though GDPR)
		log.Printf("CRIT: multiple rows returned in GetUser(email=%q): %d", email, len(users))
		return nil, fmt.Errorf("Internal Server Error")
	}
}

func (db *PostgresDB) CreateUser(u *User) error {
	err := db.pg.Insert(u)
	// TODO: what happens if unique constraint violated?
	if err != nil {
		log.Printf("CreateUser: %T %#v", err, err)
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

func (db *PostgresDB) DeleteUser(email string) error {
	// TODO: [LATER] consider using pg's "soft_delete" annotation & support
	result, err := db.pg.Model((*User)(nil)).
		Set(`deleted = ?`, time.Now()).
		Where(`email = ?`, email).
		Where(`deleted IS NULL`).
		Update()
	if err != nil {
		log.Printf("DeleteUser: %T %#v", err, err)
		return fmt.Errorf("deleting user: %w", err)
	}

	rows := result.RowsAffected()
	switch rows {
	case 0:
		// FIXME: distinct error type for 'not found'
		return fmt.Errorf("user not found: %s", email)
	case 1:
		// ok
		return nil
	default:
		log.Printf("CRIT: multiple rows affected in DeleteUser(email=%q): %d", email, rows)
		return nil
	}
}

type pgLogger struct{}

func (pgLogger) BeforeQuery(ctx context.Context, _ *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}
func (pgLogger) AfterQuery(_ context.Context, q *pg.QueryEvent) error {
	log.Println(q.FormattedQuery())
	return nil
}
