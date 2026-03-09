interface AmortizationInput {
  principal: number;
  annualRate: number;
  termYears: number;
  extraMonthly: number;
}

interface AmortizationEntry {
  month: number;
  payment: number;
  principal: number;
  interest: number;
  balance: number;
}

interface AmortizationResult {
  monthlyPayment: number;
  totalInterest: number;
  entries: AmortizationEntry[];
}

export function amortizationSchedule(input: AmortizationInput): AmortizationResult {
  const { principal, annualRate, termYears, extraMonthly } = input;

  if (principal <= 0) {
    return { monthlyPayment: 0, totalInterest: 0, entries: [] };
  }

  const monthlyRate = annualRate / 100 / 12;
  const totalMonths = termYears * 12;

  let monthlyPayment: number;
  if (monthlyRate === 0) {
    monthlyPayment = principal / totalMonths;
  } else {
    monthlyPayment =
      (principal * monthlyRate * Math.pow(1 + monthlyRate, totalMonths)) /
      (Math.pow(1 + monthlyRate, totalMonths) - 1);
  }

  const entries: AmortizationEntry[] = [];
  let balance = principal;
  let totalInterest = 0;

  for (let month = 1; month <= totalMonths && balance > 0; month++) {
    const interest = balance * monthlyRate;
    let principalPaid = monthlyPayment - interest + extraMonthly;

    if (principalPaid > balance) {
      principalPaid = balance;
    }

    balance -= principalPaid;
    totalInterest += interest;

    entries.push({
      month,
      payment: principalPaid + interest,
      principal: principalPaid,
      interest,
      balance: Math.max(0, balance),
    });
  }

  return { monthlyPayment, totalInterest, entries };
}

interface EquityInput {
  purchasePrice: number;
  downPaymentPercent: number;
  annualRate: number;
  termYears: number;
  extraMonthly: number;
  monthsElapsed: number;
  currentValuation: number;
}

interface EquityResult {
  downPayment: number;
  loanBalance: number;
  currentEquity: number;
  totalPaidToDate: number;
  principalPaidToDate: number;
  interestPaidToDate: number;
}

export function currentEquity(input: EquityInput): EquityResult {
  const downPayment = input.purchasePrice * (input.downPaymentPercent / 100);
  const loanAmount = input.purchasePrice - downPayment;

  const schedule = amortizationSchedule({
    principal: loanAmount,
    annualRate: input.annualRate,
    termYears: input.termYears,
    extraMonthly: input.extraMonthly,
  });

  const monthsCapped = Math.min(input.monthsElapsed, schedule.entries.length);
  const elapsed = schedule.entries.slice(0, monthsCapped);

  const lastEntry = elapsed[elapsed.length - 1];
  const loanBalance = lastEntry ? lastEntry.balance : loanAmount;
  const totalPaid = elapsed.reduce((sum, e) => sum + e.payment, 0);
  const principalPaid = elapsed.reduce((sum, e) => sum + e.principal, 0);
  const interestPaid = elapsed.reduce((sum, e) => sum + e.interest, 0);

  return {
    downPayment,
    loanBalance,
    currentEquity: input.currentValuation - loanBalance,
    totalPaidToDate: totalPaid,
    principalPaidToDate: principalPaid,
    interestPaidToDate: interestPaid,
  };
}

interface SafeInput {
  ownershipPercentage: number;
  valuationCap: number;
}

interface SafeScenario {
  valuation: number;
  ownershipValue: number;
  effectiveOwnership: number;
}

const SAFE_VALUATIONS = [5_000_000, 10_000_000, 25_000_000, 50_000_000, 100_000_000];

export function safeScenarios(input: SafeInput): SafeScenario[] {
  return SAFE_VALUATIONS.map((valuation) => {
    const effectiveOwnership = input.ownershipPercentage;
    const ownershipValue = valuation * (effectiveOwnership / 100);

    return {
      valuation,
      ownershipValue,
      effectiveOwnership,
    };
  });
}

interface CompoundGrowthInput {
  principal: number;
  monthlyContribution: number;
  annualRate: number;
  years: number;
}

interface CompoundGrowthResult {
  finalValue: number;
  totalContributed: number;
  totalGrowth: number;
}

export function compoundGrowth(input: CompoundGrowthInput): CompoundGrowthResult {
  const { principal, monthlyContribution, annualRate, years } = input;
  const monthlyRate = annualRate / 100 / 12;
  const totalMonths = years * 12;

  let balance = principal;
  for (let month = 0; month < totalMonths; month++) {
    balance = balance * (1 + monthlyRate) + monthlyContribution;
  }

  const totalContributed = principal + monthlyContribution * totalMonths;

  return {
    finalValue: balance,
    totalContributed,
    totalGrowth: balance - totalContributed,
  };
}
