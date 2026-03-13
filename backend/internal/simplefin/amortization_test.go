package simplefin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeAmortizationSchedule(t *testing.T) {
	tests := []struct {
		name            string
		principal       float64
		annualRate      float64
		termMonths      int
		wantPayment     float64
		wantEntries     int
		wantFinalBal    float64
		wantTotalIntGt0 bool
	}{
		{
			name:            "standard 30 year mortgage at 6.5%",
			principal:       300000,
			annualRate:      6.5,
			termMonths:      360,
			wantPayment:     1896.20,
			wantEntries:     360,
			wantFinalBal:    0,
			wantTotalIntGt0: true,
		},
		{
			name:            "15 year mortgage at 5%",
			principal:       200000,
			annualRate:      5.0,
			termMonths:      180,
			wantPayment:     1581.59,
			wantEntries:     180,
			wantFinalBal:    0,
			wantTotalIntGt0: true,
		},
		{
			name:         "zero interest rate",
			principal:    120000,
			annualRate:   0,
			termMonths:   360,
			wantPayment:  333.33,
			wantEntries:  360,
			wantFinalBal: 0,
		},
		{
			name:        "zero principal",
			principal:   0,
			annualRate:  6.5,
			termMonths:  360,
			wantPayment: 0,
			wantEntries: 0,
		},
		{
			name:            "short term 5 year loan",
			principal:       25000,
			annualRate:      4.0,
			termMonths:      60,
			wantPayment:     460.41,
			wantEntries:     60,
			wantFinalBal:    0,
			wantTotalIntGt0: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeAmortizationSchedule(tt.principal, tt.annualRate, tt.termMonths)

			assert.InDelta(t, tt.wantPayment, result.MonthlyPayment, 0.01)
			require.Len(t, result.Entries, tt.wantEntries)

			if tt.wantEntries > 0 {
				last := result.Entries[len(result.Entries)-1]
				assert.InDelta(t, tt.wantFinalBal, last.Balance, 0.01)
				assert.Equal(t, 1, result.Entries[0].Month)
				assert.Equal(t, tt.wantEntries, last.Month)
			}

			if tt.wantTotalIntGt0 {
				assert.Greater(t, result.TotalInterest, 0.0)
			}
		})
	}
}

func TestAmortizationSchedule_BalanceDecreases(t *testing.T) {
	result := ComputeAmortizationSchedule(250000, 6.0, 360)

	for i := 1; i < len(result.Entries); i++ {
		assert.LessOrEqual(t, result.Entries[i].Balance, result.Entries[i-1].Balance,
			"balance should decrease monotonically at month %d", result.Entries[i].Month)
	}
}

func TestAmortizationSchedule_PrincipalAndInterestSumToPayment(t *testing.T) {
	result := ComputeAmortizationSchedule(200000, 5.5, 360)

	for _, entry := range result.Entries {
		assert.InDelta(t, entry.Payment, entry.Principal+entry.Interest, 0.01,
			"payment should equal principal + interest at month %d", entry.Month)
	}
}
