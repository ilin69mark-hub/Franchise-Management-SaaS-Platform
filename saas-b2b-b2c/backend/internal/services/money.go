package services

import (
	"github.com/shopspring/decimal"
)

// money.go — точные денежные расчёты через decimal.
//
// Правило: деньги в БД хранятся как NUMERIC, в Go-сущностях — decimal.Decimal.
// На границу API (JSON) отдаём float64 через InexactFloat64(), т.к. фронт
// ожидает number. Вся арифметика в Go — только через decimal с округлением
// до копеек (Round(2)), чтобы избежать ошибок 0.1+0.2.

// MoneyFromFloat конвертирует float в decimal с округлением до копеек.
func MoneyFromFloat(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f).Round(2)
}

// MoneySub вычитание с округлением до копеек, результат как float для API.
func MoneySub(a, b float64) float64 {
	return decimal.NewFromFloat(a).Sub(decimal.NewFromFloat(b)).Round(2).InexactFloat64()
}

// MoneyDiv делит сумму на количество (средний чек) с округлением до копеек.
func MoneyDiv(sum float64, count int64) float64 {
	if count <= 0 {
		return 0
	}
	return decimal.NewFromFloat(sum).Div(decimal.NewFromInt(count)).Round(2).InexactFloat64()
}

// MoneyPercent считает (part/total*100) как int-процент через decimal.
func MoneyPercent(part, total float64) int {
	if total <= 0 {
		return 0
	}
	p := decimal.NewFromFloat(part).Div(decimal.NewFromFloat(total)).Mul(decimal.NewFromInt(100)).Round(0)
	return int(p.InexactFloat64())
}
