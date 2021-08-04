package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

type controller struct {
	service Service
}

func RegisterHandlers(api fiber.Router, service Service) {
	c := controller{service}

	// Assign handlers to endpoints
	api.Get("/currency", c.GetCurrency)
	api.Get("/currency_rate", c.GetAccountCurrencyRate)
	api.Get("/balance", c.GetBalance)
	api.Post("/add", c.AddFunds)
	api.Post("/withdraw", c.Withdraw)
}

// Represents right request json body for `add` and `withdraw`,
// otherwise client will get 400 code and `errorResponse`
type amountRequest struct {
	Amount float64 `json:"amount"`
}

// Represents right request json body for `balance` and `currency_rate`,
// otherwise client will get 400 code and `errorResponse`
type currencyRequest struct {
	Currency string `json:"currency"`
}

// Represents response of `balance`, `add` and `withdraw` endpoints when everything is successful
type balanceResponse struct {
	Balance json.Number `json:"balance"`
}

// If any error occurs in endpoint, this struct will be returned to client
type errorResponse struct {
	Error string `json:"error"`
}

// Withdraw Проверяет в запросе поле "amount" и вносит на счёт сумму `amount`,
// при успехе отдаёт новый баланс "balance"
func (c controller) AddFunds(ctx *fiber.Ctx) error {
	var req amountRequest
	err := ctx.BodyParser(&req)
	if err != nil {
		return ctx.Status(400).JSON(errorResponse{requiredField("amount")})
	}

	c.service.AddFunds(req.Amount)
	c.service.SumProfit()

	balance := c.service.GetBalance(CurrencySBP)
	log.Printf("Add %.2f, Balance is %.2f\n", req.Amount, balance)
	return ctx.JSON(balanceResponse{toJSNumber(balance, 2)})
}

// Withdraw Проверяет в запросе поле "amount" и пытается списать средства со счёта,
// при успехе отдаёт новый баланс "balance"
func (c controller) Withdraw(ctx *fiber.Ctx) error {
	var req amountRequest
	err := ctx.BodyParser(&req)
	if err != nil {
		return ctx.Status(400).JSON(errorResponse{requiredField("amount")})
	}

	err = c.service.Withdraw(req.Amount)
	if err != nil {
		if err == ErrWithdrawCondition {
			log.Printf("Tried to withdraw %.2f, but condition is false", req.Amount)
		}
		// If it's a real error from SQL, it likely to report there
		return ctx.Status(200).JSON(errorResponse{err.Error()})
	}

	balance := c.service.GetBalance(CurrencySBP)

	log.Printf("Withdrawed %.2f, Balance is %.2f\n", req.Amount, balance)

	return ctx.JSON(balanceResponse{toJSNumber(balance, 2)})
}

// GetCurrency Возвращает валюту "currency" в котором открыт счёт
func (c controller) GetCurrency(ctx *fiber.Ctx) error {
	currency := c.service.GetCurrency()

	log.Printf("Account currency is %s\n", currency)

	return ctx.JSON(struct {
		Currency string `json:"currency"`
	}{
		Currency: string(currency),
	})
}

// GetAccountCurrencyRate Проверяет в запросе поле "currency" на валидность и отдаёт соот. курс "currencyRate" из валюты счёта
func (c controller) GetAccountCurrencyRate(ctx *fiber.Ctx) error {
	var req currencyRequest
	err := ctx.BodyParser(&req)
	if err != nil {
		return ctx.Status(400).JSON(errorResponse{requiredField("currency")})
	}

	currency, err := req2Currency(req.Currency)
	if err != nil {
		return ctx.Status(200).JSON(errorResponse{err.Error()})
	}

	cr := c.service.GetAccountCurrencyRate(currency)

	log.Printf("Currency rate is %.4f\n", cr)

	return ctx.JSON(struct {
		CurrencyRate json.Number `json:"currencyRate"`
	}{
		CurrencyRate: toJSNumber(cr, 4),
	})
}

// GetBalance Проверяет в запросе поле "currency" на валидность и отдаёт баланс "balance" в соот. валюте
func (c controller) GetBalance(ctx *fiber.Ctx) error {
	var req currencyRequest
	err := ctx.BodyParser(&req)
	if err != nil {
		return ctx.Status(400).JSON(errorResponse{requiredField("currency")})
	}

	currency, err := req2Currency(req.Currency)
	if err != nil {
		return ctx.Status(200).JSON(errorResponse{err.Error()})
	}

	balance := c.service.GetBalance(currency)

	log.Printf("Balance is %.2f\n", balance)

	return ctx.JSON(balanceResponse{toJSNumber(balance, 2)})
}

// Casts text into Currency and if it doesn't match with "enum" values of Currency returns error
func req2Currency(text string) (Currency, error) {
	currency := Currency(text)
	switch currency {
	case CurrencyRUB, CurrencySBP:
		return currency, nil
	default:
		return "", errors.New(`currency is not valid, there is only SBP and RUB`)
	}
}

// Adds "required field " before `field`
func requiredField(field string) string {
	return fmt.Sprintf(`required field "%s"`, field)
}

// Formatting float with `precision` and wraps it into `json.Number`
func toJSNumber(f float64, precision uint) json.Number {
	return json.Number(fmt.Sprintf("%.*f", precision, f))
}
