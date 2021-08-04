package account

import (
	"database/sql"
	"embed"
	"fmt"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

const (
	// ID нашего единственного вкладчика
	theOnlyLoyalDepositorID = "1"
)

// We will pretend that errors do not exist
type Repository interface {
	// Gets account of the only loyal depositor of our bank
	GetAccount() Account
	// Updates account of the only loyal depositor of our bank
	UpdateAccount(account Account)
}

type repo struct {
	// We're old :(
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return repo{db}
}

// Gets balance of the only loyal depositor of our bank
func (r repo) GetAccount() Account {
	var balance int64
	var currency Currency
	err := r.db.QueryRow("SELECT balance, currency FROM accounts WHERE id = ?", theOnlyLoyalDepositorID).Scan(&balance, &currency)
	if err != nil {
		panic("the only depositor is... not there")
	}

	return Account{balance, currency}
}

// Updates account of the only loyal depositor of our bank
func (r repo) UpdateAccount(account Account) {
	result, err := r.db.Exec("UPDATE accounts SET balance = ?, currency = ? WHERE id = ?", account.balance, account.currency, theOnlyLoyalDepositorID)
	if err != nil {
		panic(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}

	if rows != 1 {
		panic(fmt.Sprintf("expected to affect 1 row, affected %d", rows))
	}
}

//go:embed migrations
var migrations embed.FS

func LoadMigrations(db *sql.DB) error {
	// Мне так понравилось как всё легко сложилось с миграциями в sqlite
	src, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return err
	}

	// Вот бы везде так было
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance(
		"httpfs", src,
		"sqlite3", driver,
	)
	if err != nil {
		return err
	}

	// Applies all up migrations from current version
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
