package account

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

func RegisterHandlers(api fiber.Router, service Service) {
	c := controller{service}

	api.Get("/currency", c.GetCurrency)
	api.Get("/currency_rate", c.GetAccountCurrencyRate)
	api.Get("/balance", c.GetBalance)
	api.Post("/add", c.AddFunds)
	api.Post("/withdraw", c.Withdraw)
}

type controller struct {
	service Service
}

func (c controller) AddFunds(ctx *fiber.Ctx) error {
	var req struct {
		Amount float64 `json:"amount"`
	}
	err := ctx.BodyParser(&req)
	if err != nil {
		return ctx.Status(400).JSON(struct {
			Error string `json:"error"`
		}{
			Error: requiredField("amount"),
		})
	}

	c.service.AddFunds(req.Amount)
	c.service.SumProfit()

	balance := c.service.GetBalance(CurrencySBP)
	log.Printf("Add %.2f, Balance is %.2f\n", req.Amount, balance)
	return ctx.JSON(struct {
		Balance json.Number `json:"balance"`
	}{
		Balance: toJSNumber(balance, 2),
	})
}

func (c controller) Withdraw(ctx *fiber.Ctx) error {
	var req struct {
		Amount float64 `json:"amount"`
	}
	err := ctx.BodyParser(&req)
	if err != nil {
		return ctx.Status(400).JSON(struct {
			Error string `json:"error"`
		}{
			Error: requiredField("amount"),
		})
	}

	err = c.service.Withdraw(req.Amount)
	if err != nil {
		if err == ErrWithdrawCondition {
			log.Printf("Tried to withdraw %.2f, but condition is false", req.Amount)
		}

		return ctx.Status(200).JSON(struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
	}

	balance := c.service.GetBalance(CurrencySBP)
	log.Printf("Withdrawed %.2f, Balance is %.2f\n", req.Amount, balance)
	return ctx.JSON(struct {
		Balance json.Number `json:"balance"`
	}{
		Balance: toJSNumber(balance, 2),
	})
}

func (c controller) GetCurrency(ctx *fiber.Ctx) error {
	currrency := c.service.GetCurrency()
	log.Printf("Account currency is %s\n", currrency)
	return ctx.JSON(struct {
		Currency string `json:"currency"`
	}{
		Currency: string(currrency),
	})
}

func (c controller) GetAccountCurrencyRate(ctx *fiber.Ctx) error {
	var req struct {
		Currency string `json:"currency"`
	}
	err := ctx.BodyParser(&req)
	if err != nil {
		return ctx.Status(400).JSON(struct {
			Error string `json:"error"`
		}{
			Error: requiredField("currency"),
		})
	}

	currency := Currency(req.Currency)
	switch currency {
	case CurrencyRUB, CurrencySBP:
		break
	default:
		return ctx.Status(400).JSON(struct {
			Error string `json:"error"`
		}{
			Error: requiredField("currency"),
		})
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
	var req struct {
		Currency string `json:"currency"`
	}
	err := ctx.BodyParser(&req)
	if err != nil {
		return ctx.Status(400).JSON(struct {
			Error string `json:"error"`
		}{
			Error: requiredField("currency"),
		})
	}

	balance := c.service.GetBalance(Currency(req.Currency))
	log.Printf("Balance is %.2f\n", balance)
	return ctx.JSON(struct {
		Balance json.Number `json:"balance"`
	}{
		Balance: toJSNumber(balance, 2),
	})
}

func requiredField(field string) string {
	return fmt.Sprintf(`required field "%s"`, field)
}

func toJSNumber(f float64, precision uint) json.Number {
	return json.Number(fmt.Sprintf("%.*f", precision, f))
}
