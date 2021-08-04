package account

import (
	"errors"
	"math"
)

type Currency string

const (
	CurrencySBP Currency = "SBP"
	CurrencyRUB Currency = "RUB"
)

type Account struct {
	balance  int64
	currency Currency
}

const (
	fixedFraction = 2
)

// Returns currency of Account
func (acc Account) Currency() Currency {
	return acc.currency
}

// Returns balance of account in `currency`
func (acc Account) Amount(currency Currency) (float64, error) {
	switch currency {
	case CurrencySBP, CurrencyRUB:
		break
	default:
		return 0, errors.New("not valid currency")
	}

	if acc.currency == currency {
		return float64(acc.balance) / math.Pow10(fixedFraction), nil
	} else if acc.currency != CurrencyRUB {
		return float64(acc.balance) * exchangeRateSBP2RUB / math.Pow10(fixedFraction), nil
	}

	panic("unreachable")
}

// Add to account balance `amount` SBP
func (acc Account) Add(amount float64) Account {
	return Account{acc.balance + int64(math.Ceil(amount*math.Pow10(fixedFraction))), acc.currency}
}

// Subtract account balance `amount` SBP
func (acc Account) Sub(amount float64) Account {
	return acc.Add(-amount)
}
