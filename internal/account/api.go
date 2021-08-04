package account

import (
	"encoding/json"
	"fmt"

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

// If any error occurs in endpoint, this struct will be returned to client
type errorResponse struct {
	Error string `json:"error"`
}

// Withdraw Проверяет в запросе поле "amount" и вносит на счёт сумму `amount`,
// при успехе отдаёт новый баланс "balance"
func (c controller) AddFunds(ctx *fiber.Ctx) error {
	var req amountRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(errorResponse{requiredField("amount")})
	}

	c.service.AddFunds(req.Amount)
	c.service.SumProfit()

	balance, err := c.service.GetBalance(string(CurrencySBP))
	if err != nil {
		return ctx.JSON(errorResponse{err.Error()})
	}

	return ctx.JSON(fiber.Map{
		"balance": toJSNumber(balance, 2),
	})
}

// Withdraw Проверяет в запросе поле "amount" и пытается списать средства со счёта,
// при успехе отдаёт новый баланс "balance"
func (c controller) Withdraw(ctx *fiber.Ctx) error {
	var req amountRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(errorResponse{requiredField("amount")})
	}

	if err := c.service.Withdraw(req.Amount); err != nil {
		return ctx.Status(200).JSON(errorResponse{err.Error()})
	}

	balance, err := c.service.GetBalance(string(CurrencySBP))
	if err != nil {
		return ctx.JSON(errorResponse{err.Error()})
	}

	return ctx.JSON(fiber.Map{
		"balance": toJSNumber(balance, 2),
	})
}

// GetCurrency Возвращает валюту "currency" в котором открыт счёт
func (c controller) GetCurrency(ctx *fiber.Ctx) error {
	currency, err := c.service.GetCurrency()
	if err != nil {
		return ctx.JSON(errorResponse{err.Error()})
	}
	return ctx.JSON(fiber.Map{
		"currency": currency,
	})
}

// GetAccountCurrencyRate Проверяет в запросе поле "currency" на валидность и отдаёт соот. курс "currencyRate" из валюты счёта
func (c controller) GetAccountCurrencyRate(ctx *fiber.Ctx) error {
	var req currencyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(errorResponse{requiredField("currency")})
	}

	rate, err := c.service.GetAccountCurrencyRate(req.Currency)
	if err != nil {
		return ctx.JSON(errorResponse{err.Error()})
	}

	return ctx.JSON(fiber.Map{
		"currencyRate": toJSNumber(rate, 4),
	})
}

// GetBalance Проверяет в запросе поле "currency" на валидность и отдаёт баланс "balance" в соот. валюте
func (c controller) GetBalance(ctx *fiber.Ctx) error {
	var req currencyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(errorResponse{requiredField("currency")})
	}

	balance, err := c.service.GetBalance(req.Currency)
	if err != nil {
		return ctx.JSON(errorResponse{err.Error()})
	}

	return ctx.JSON(fiber.Map{
		"balance": toJSNumber(balance, 2),
	})
}

// Adds "required field " before `field`
func requiredField(field string) string {
	return fmt.Sprintf(`required field "%s"`, field)
}

// Formatting float with `precision` and wraps it into `json.Number`
func toJSNumber(f float64, precision uint) json.Number {
	return json.Number(fmt.Sprintf("%.*f", precision, f))
}
