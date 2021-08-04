package account

import (
	"errors"
	"fmt"
)

const (
	// Процент по вкладу
	investmentMultiplier = 0.06

	// Максимальная часть счёта к выводу
	maxWithdrawPercent = 0.3

	exchangeRateSBP2RUB float64 = 0.7523
)

type Service interface {
	// AddFunds Позволяет внести на счёт сумму sum
	AddFunds(sum float64) error
	// SumProfit Рассчитывает процент по вкладу и полученные деньги вносит на счёт
	SumProfit() error
	// Withdraw Производит списание со счёта по указанным правилам. Если списание выходит за рамки правил, выдаёт ошибку
	Withdraw(sum float64) error
	// GetCurrency Выдаёт валюту счёта
	GetCurrency() (string, error)
	// GetAccountCurrencyRate Выдаёт курс валюты счёта к передаваемой валюте cur
	GetAccountCurrencyRate(cur string) (float64, error)
	// GetBalance Выдаёт баланс счёта в указанной валюте
	GetBalance(cur string) (float64, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo}
}

// AddFunds Позволяет внести на счёт сумму `amount`
func (s service) AddFunds(sum float64) error {
	account, err := s.repo.GetAccount()
	if err != nil {
		return err
	}
	return s.repo.UpdateAccount(account.Add(sum))
}

// SumProfit Рассчитывает процент по вкладу и полученные деньги вносит на счёт
func (s service) SumProfit() error {
	account, err := s.repo.GetAccount()
	if err != nil {
		return err
	}

	balance, err := account.Amount(CurrencySBP)
	if err != nil {
		return err
	}

	return s.repo.UpdateAccount(account.Add(balance * investmentMultiplier))
}

// Withdraw Производит списание со счёта по указанным правилам. Если списание выходит за рамки правил, выдаёт ошибку
func (s service) Withdraw(sum float64) error {
	account, err := s.repo.GetAccount()
	if err != nil {
		return err
	}

	balance, err := account.Amount(CurrencySBP)
	if err != nil {
		return err
	}

	// Only if amount of withdraw less that 30% of balance, it can proceed
	if balance*maxWithdrawPercent < sum {
		return ErrWithdrawCondition
	}

	return s.repo.UpdateAccount(account.Sub(sum))
}

// GetCurrency Выдаёт валюту счёта
func (s service) GetCurrency() (string, error) {
	account, err := s.repo.GetAccount()
	if err != nil {
		return "", err
	}
	return string(account.Currency()), nil
}

// GetAccountCurrencyRate Выдаёт курс валюты счёта к передаваемой валюте
func (s service) GetAccountCurrencyRate(cur string) (float64, error) {
	switch Currency(cur) {
	case CurrencySBP:
		return 1, nil
	case CurrencyRUB:
		return exchangeRateSBP2RUB, nil
	default:
		return 0, errors.New("not valid currency")
	}
}

// GetBalance Выдаёт баланс счёта в указанной валюте
func (s service) GetBalance(cur string) (float64, error) {
	account, err := s.repo.GetAccount()
	if err != nil {
		return 0, err
	}

	balance, err := account.Amount(Currency(cur))
	if err != nil {
		return 0, err
	}

	return balance, nil
}

var ErrWithdrawCondition error = fmt.Errorf("the amount of withdrawal exceeds the allowable %.2f%% of balance", maxWithdrawPercent*100)
