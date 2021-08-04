package account

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

const (
	// ID нашего единственного вкладчика
	theOnlyLoyalDepositorID = "1"

	// Процент по вкладу
	investmentPercentageIncrease = 1 + 0.06

	// Максимальная часть счёта к выводу
	maxWithdrawPercent = 0.3
)

type Currency string

const (
	CurrencySBP Currency = "SBP"
	CurrencyRUB Currency = "RUB"
)

var exchangeRates = map[string]float64{
	exchangeRateKey(CurrencySBP, CurrencyRUB): 0.7523,
	exchangeRateKey(CurrencyRUB, CurrencySBP): 1.3292,
}

type Service interface {
	// AddFunds Позволяет внести на счёт сумму sum
	AddFunds(sum float64)
	// SumProfit Рассчитывает процент по вкладу и полученные деньги вносит на счёт
	SumProfit()
	// Withdraw Производит списание со счёта по указанным правилам. Если списание выходит за рамки правил, выдаёт ошибку
	Withdraw(f float64) error
	// GetCurrency Выдаёт валюту счёта
	GetCurrency() Currency
	// GetAccountCurrencyRate Выдаёт курс валюты счёта к передаваемой валюте cur
	GetAccountCurrencyRate(cur Currency) float64
	// GetBalance Выдаёт баланс счёта в указанной валюте
	GetBalance(cur Currency) float64
}

type service struct {
	// We're old :(
	db *sql.DB
}

func NewService(db *sql.DB) Service {
	return service{db}
}

func (s service) AddFunds(sum float64) {
	balance, err := s.getBalance()
	if err != nil {
		log.Println(err)
		return
	}

	// I know that we can add sum with sql, it's just the right way, imo, to have business logic on service side
	balance += sum

	err = s.setBalance(balance)
	if err != nil {
		log.Println(err)
		return
	}
}

func (s service) SumProfit() {
	balance, err := s.getBalance()
	if err != nil {
		log.Println(err)
		return
	}

	balance *= investmentPercentageIncrease

	err = s.setBalance(balance)
	if err != nil {
		log.Println(err)
		return
	}
}

func (s service) Withdraw(amount float64) error {
	balance, err := s.getBalance()
	if err != nil {
		return err
	}

	if amount > maxWithdrawPercent*balance {
		return ErrWithdrawCondition
	}
	balance -= amount

	err = s.setBalance(balance)
	if err != nil {
		return err
	}

	return nil
}

func (s service) GetCurrency() Currency {
	var currency string
	err := s.db.QueryRow("SELECT currency FROM accounts WHERE id = ?", theOnlyLoyalDepositorID).Scan(&currency)
	if err != nil {
		return ""
	}

	return Currency(currency)
}

func (s service) GetAccountCurrencyRate(cur Currency) float64 {
	return exchangeRate(CurrencySBP, cur)
}

func (s service) GetBalance(cur Currency) float64 {
	balance, err := s.getBalance()
	if err != nil {
		log.Println(err)
	}

	return balance * exchangeRate(CurrencySBP, cur)
}

// Internal method for getting balance
func (s service) getBalance() (float64, error) {
	var balance float64
	err := s.db.QueryRow("SELECT balance FROM accounts WHERE id = ?", theOnlyLoyalDepositorID).Scan(&balance)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

// Internal method for setting balance
func (s service) setBalance(balance float64) error {
	result, err := s.db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", balance, theOnlyLoyalDepositorID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}

	return nil
}

func exchangeRate(from Currency, to Currency) float64 {
	if from == to {
		return 1
	}
	return exchangeRates[exchangeRateKey(from, to)]
}

const (
	exchangeRateKeyDelimiter = "->"
)

// Util function for getting exchange rates map key
func exchangeRateKey(from Currency, to Currency) string {
	return string(from) + exchangeRateKeyDelimiter + string(to)
}

var ErrWithdrawCondition error = fmt.Errorf("the amount of withdrawal exceeds the allowable %.2f%% amount", maxWithdrawPercent*100)

//go:embed migrations
var migrations embed.FS

func LoadMigrations(db *sql.DB) error {
	src, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return err
	}

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
