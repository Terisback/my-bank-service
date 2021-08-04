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

	api.Get("/currency", c.GetCurrency)
	api.Get("/currency_rate", c.GetAccountCurrencyRate)
	api.Get("/balance", c.GetBalance)
	api.Post("/add", c.AddFunds)
	api.Post("/withdraw", c.Withdraw)
}


type amountRequest struct {
	Amount float64 `json:"amount"`
}

type currencyRequest struct {
	Currency string `json:"currency"`
}

type balanceResponse struct {
	Balance json.Number `json:"balance"`
}

type errorResponse struct {
	Error string `json:"error"`
}

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

		return ctx.Status(200).JSON(errorResponse{err.Error()})
	}

	balance := c.service.GetBalance(CurrencySBP)
	log.Printf("Withdrawed %.2f, Balance is %.2f\n", req.Amount, balance)
	return ctx.JSON(balanceResponse{toJSNumber(balance, 2)})
}

func (c controller) GetCurrency(ctx *fiber.Ctx) error {
	currency := c.service.GetCurrency()

	log.Printf("Account currency is %s\n", currency)

	return ctx.JSON(struct {
		Currency string `json:"currency"`
	}{
		Currency: string(currency),
	})
}

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

func req2Currency(text string) (Currency, error) {
	currency := Currency(text)
	switch currency {
	case CurrencyRUB, CurrencySBP:
		return currency, nil
	default:
		return "", errors.New(`currency is not valid, there is only SBP and RUB`)
	}
}

func requiredField(field string) string {
	return fmt.Sprintf(`required field "%s"`, field)
}

func toJSNumber(f float64, precision uint) json.Number {
	return json.Number(fmt.Sprintf("%.*f", precision, f))
}
