package simplefin

import "math"

type AmortizationEntry struct {
	Month     int
	Payment   float64
	Principal float64
	Interest  float64
	Balance   float64
}

type AmortizationResult struct {
	MonthlyPayment float64
	TotalInterest  float64
	Entries        []AmortizationEntry
}

func ComputeAmortizationSchedule(principal, annualRate float64, termMonths int) AmortizationResult {
	if principal <= 0 {
		return AmortizationResult{}
	}

	monthlyRate := annualRate / 100 / 12

	var monthlyPayment float64
	if monthlyRate == 0 {
		monthlyPayment = principal / float64(termMonths)
	} else {
		pow := math.Pow(1+monthlyRate, float64(termMonths))
		monthlyPayment = (principal * monthlyRate * pow) / (pow - 1)
	}

	entries := make([]AmortizationEntry, 0, termMonths)
	balance := principal
	totalInterest := 0.0

	for month := 1; month <= termMonths && balance > 0; month++ {
		interest := balance * monthlyRate
		principalPaid := monthlyPayment - interest

		if principalPaid > balance {
			principalPaid = balance
		}

		balance -= principalPaid
		totalInterest += interest

		entries = append(entries, AmortizationEntry{
			Month:     month,
			Payment:   principalPaid + interest,
			Principal: principalPaid,
			Interest:  interest,
			Balance:   math.Max(0, balance),
		})
	}

	return AmortizationResult{
		MonthlyPayment: monthlyPayment,
		TotalInterest:  totalInterest,
		Entries:        entries,
	}
}
