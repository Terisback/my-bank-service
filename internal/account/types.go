package account

import "math"

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
func (acc Account) Amount(currency Currency) float64 {
	switch currency {
	case CurrencySBP, CurrencyRUB:
		break
	default:
		panic("not valid currency")
	}

	if acc.currency == currency {
		return float64(acc.balance) / math.Pow10(fixedFraction)
	} else if acc.currency != CurrencyRUB {
		return float64(acc.balance) * exchangeRateSBP2RUB / math.Pow10(fixedFraction)
	}

	panic("unreachable")
}

// Add to account balance `amount` SBP
func (acc Account) Add(amount float64) Account {
	return Account{acc.balance + int64(amount*math.Pow10(fixedFraction)), acc.currency}
}

// Subtract account balance `amount` SBP
func (acc Account) Sub(amount float64) Account {
	return Account{acc.balance - int64(amount*math.Pow10(fixedFraction)), acc.currency}
}

// Multiply account balance by `m`
func (acc Account) Mul(m float64) Account {
	return Account{acc.balance - int64(m*math.Pow10(fixedFraction)), acc.currency}
}
