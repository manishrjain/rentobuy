import type { CalculatorInputs, CalculationResults, Period } from '../types';

export function calculateMonthlyPayment(
  principal: number,
  monthlyRate: number,
  months: number
): number {
  if (monthlyRate === 0) {
    return principal / months;
  }
  const factor = Math.pow(1 + monthlyRate, months);
  return (principal * (monthlyRate * factor)) / (factor - 1);
}

export function getPeriods(loanDuration: number, projectionYears: number): Period[] {
  const maxYears = projectionYears;
  const maxMonths = maxYears * 12;

  const periods: Period[] = [];

  // Generate periods for years 0 to maxYears
  for (let year = 0; year <= maxYears; year++) {
    const months = year * 12;
    const loanTermYear = loanDuration / 12;

    // Check if this is the loan term year (and loan term is not a round year already covered)
    if (loanDuration > 0 && loanDuration === months && loanDuration % 12 === 0 && loanDuration <= maxMonths) {
      // Mark this year as the loan term year
      periods.push({ label: `X${year.toString().padStart(3)}y`, months });
    } else {
      periods.push({ label: `${year.toString().padStart(3)}y`, months });
    }
  }

  return periods;
}

// Calculate effective loan values accounting for any elapsed time
export function getEffectiveLoanValues(inputs: CalculatorInputs): {
  effectiveLoanAmount: number;
  effectiveLoanTerm: number;
  monthlyLoanPayment: number;
  refinanceCashOut: number;
} {
  if (inputs.loanAmount <= 0) {
    return { effectiveLoanAmount: 0, effectiveLoanTerm: 0, monthlyLoanPayment: 0, refinanceCashOut: 0 };
  }

  const monthlyRate = inputs.loanRate / 100 / 12;

  // If refinance is enabled, use refinance values directly
  if (inputs.includeRefinance) {
    const payment = calculateMonthlyPayment(inputs.loanAmount, monthlyRate, inputs.loanTerm);
    const payoffBalance = inputs.payoffBalance ?? 0;
    const closingCosts = inputs.closingCosts ?? 0;
    const refinanceCashOut = inputs.loanAmount - payoffBalance - closingCosts;

    return {
      effectiveLoanAmount: inputs.loanAmount,
      effectiveLoanTerm: inputs.loanTerm,
      monthlyLoanPayment: payment,
      refinanceCashOut,
    };
  }

  const remainingTerm = inputs.remainingLoanTerm ?? inputs.loanTerm;

  // If no elapsed time, use original values
  if (remainingTerm >= inputs.loanTerm) {
    const payment = calculateMonthlyPayment(inputs.loanAmount, monthlyRate, inputs.loanTerm);
    return {
      effectiveLoanAmount: inputs.loanAmount,
      effectiveLoanTerm: inputs.loanTerm,
      monthlyLoanPayment: payment,
      refinanceCashOut: 0,
    };
  }

  // Calculate remaining balance based on elapsed time
  const originalPayment = calculateMonthlyPayment(inputs.loanAmount, monthlyRate, inputs.loanTerm);
  const monthsElapsed = inputs.loanTerm - remainingTerm;

  let balance = inputs.loanAmount;
  for (let i = 0; i < monthsElapsed; i++) {
    const interestPayment = balance * monthlyRate;
    const principalPayment = originalPayment - interestPayment;
    balance -= principalPayment;
  }

  // Recalculate payment based on remaining balance and term
  const newPayment = calculateMonthlyPayment(balance, monthlyRate, remainingTerm);

  return {
    effectiveLoanAmount: balance,
    effectiveLoanTerm: remainingTerm,
    monthlyLoanPayment: newPayment,
    refinanceCashOut: 0,
  };
}

export function populateMonthlyCosts(inputs: CalculatorInputs): {
  monthlyBuyingCosts: number[];
  monthlyRentingCosts: number[];
  remainingLoanBalance: number[];
  cumulativePrincipalPaid: number[];
  cumulativeInterestPaid: number[];
} {
  const maxMonths = 360;
  const monthlyBuyingCosts: number[] = new Array(maxMonths);
  const monthlyRentingCosts: number[] = new Array(maxMonths);
  const remainingLoanBalance: number[] = new Array(maxMonths);
  const cumulativePrincipalPaid: number[] = new Array(maxMonths);
  const cumulativeInterestPaid: number[] = new Array(maxMonths);

  const totalAnnualExpenses = inputs.annualInsurance + inputs.annualTaxes;
  let monthlyRecurringExpenses = totalAnnualExpenses / 12 - inputs.annualIncome / 12;

  const monthlyRate = inputs.loanRate / 100 / 12;
  const { effectiveLoanAmount, effectiveLoanTerm, monthlyLoanPayment } = getEffectiveLoanValues(inputs);

  // Mortgage interest deduction rate
  const taxDeductionRate = inputs.mortgageInterestDeduction / 100;

  let currentRentingCost =
    inputs.monthlyRent + inputs.annualRentCosts / 12 + inputs.otherAnnualCosts / 12;
  let currentRecurringExpenses = monthlyRecurringExpenses;

  let currentBalance = effectiveLoanAmount;
  let totalPrincipalPaid = 0;
  let totalInterestPaid = 0;

  for (let i = 0; i < maxMonths; i++) {
    if (i > 0 && i % 12 === 0) {
      currentRentingCost *= 1 + inputs.inflationRate / 100;
      currentRecurringExpenses *= 1 + inputs.inflationRate / 100;
    }

    monthlyRentingCosts[i] = currentRentingCost;

    if (i < effectiveLoanTerm) {
      const interestPayment = currentBalance * monthlyRate;
      const principalPayment = monthlyLoanPayment - interestPayment;

      // Apply tax deduction to reduce effective interest cost
      const effectiveInterestPayment = interestPayment * (1 - taxDeductionRate);
      monthlyBuyingCosts[i] = principalPayment + effectiveInterestPayment + currentRecurringExpenses;

      currentBalance -= principalPayment;

      totalPrincipalPaid += principalPayment;
      totalInterestPaid += interestPayment; // Track actual interest paid (pre-deduction)

      remainingLoanBalance[i] = currentBalance;
      cumulativePrincipalPaid[i] = totalPrincipalPaid;
      cumulativeInterestPaid[i] = totalInterestPaid;
    } else {
      monthlyBuyingCosts[i] = currentRecurringExpenses;
      remainingLoanBalance[i] = 0;
      cumulativePrincipalPaid[i] = totalPrincipalPaid;
      cumulativeInterestPaid[i] = totalInterestPaid;
    }
  }

  return {
    monthlyBuyingCosts,
    monthlyRentingCosts,
    remainingLoanBalance,
    cumulativePrincipalPaid,
    cumulativeInterestPaid,
  };
}

export function calculateKeepInvestmentTracking(
  monthlyBuyingCosts: number[],
  investmentReturnRate: number,
  maxMonths: number,
  initialCashOut: number = 0
): {
  monthlyKeepNetPosition: number[];
  monthlyKeepInvestmentReturns: number[];
} {
  const monthlyKeepNetPosition: number[] = new Array(maxMonths);
  const monthlyKeepInvestmentReturns: number[] = new Array(maxMonths);

  // Net position: positive = invested cash, negative = debt/deficit
  // Only positive net position earns investment returns
  let netPosition = initialCashOut;
  let totalReturns = 0;
  const monthlyInvestmentRate = investmentReturnRate / 100 / 12;

  for (let i = 0; i < maxMonths; i++) {
    const monthlyCost = monthlyBuyingCosts[i];

    // Negative cost = income (positive cash flow), positive cost = expense (negative cash flow)
    // Cash flow is applied directly to net position
    netPosition -= monthlyCost;

    // Only earn returns if net position is positive
    if (netPosition > 0) {
      const monthlyReturn = netPosition * monthlyInvestmentRate;
      totalReturns += monthlyReturn;
      netPosition *= 1 + monthlyInvestmentRate;
    }

    monthlyKeepNetPosition[i] = netPosition;
    monthlyKeepInvestmentReturns[i] = totalReturns;
  }

  return {
    monthlyKeepNetPosition,
    monthlyKeepInvestmentReturns,
  };
}

export function calculateAssetValue(
  startingPrice: number,
  months: number,
  appreciationRates: number[]
): number {
  let assetValue = startingPrice;
  const years = Math.floor(months / 12);
  const remainingMonths = months % 12;

  for (let year = 0; year < years; year++) {
    const rateIndex = Math.min(year, appreciationRates.length - 1);
    assetValue *= 1 + appreciationRates[rateIndex] / 100;
  }

  if (remainingMonths > 0) {
    const rateIndex = Math.min(years, appreciationRates.length - 1);
    const partialYearFactor = Math.pow(
      1 + appreciationRates[rateIndex] / 100,
      remainingMonths / 12
    );
    assetValue *= partialYearFactor;
  }

  return assetValue;
}

export function calculateSaleProceeds(
  inputs: CalculatorInputs,
  months: number,
  remainingLoanBalance: number[],
  effectiveLoanAmount?: number,
  includeSelling: boolean = true
): {
  salePrice: number;
  totalSellingCosts: number;
  loanPayoff: number;
  capitalGains: number;
  taxOnGains: number;
  netProceeds: number;
} {
  // For buy_vs_rent, always use purchasePrice as starting point
  // For sell_vs_keep, use currentMarketValue if provided (property may have appreciated)
  const startingPrice = inputs.scenario === 'sell_vs_keep' && inputs.currentMarketValue
    ? inputs.currentMarketValue
    : inputs.purchasePrice;
  const salePrice = calculateAssetValue(startingPrice, months, inputs.appreciationRate);

  let loanPayoff: number;
  if (months === 0 && effectiveLoanAmount !== undefined) {
    // At month 0, use the effective loan amount directly
    loanPayoff = effectiveLoanAmount;
  } else if (months === 0) {
    // Fallback: estimate from first month's balance
    const firstMonthBalance = remainingLoanBalance[0];
    const secondMonthBalance = remainingLoanBalance[1];
    const firstPrincipalPayment = firstMonthBalance - secondMonthBalance;
    loanPayoff = firstMonthBalance + firstPrincipalPayment;
  } else {
    const monthIndex = Math.min(months - 1, remainingLoanBalance.length - 1);
    loanPayoff = remainingLoanBalance[monthIndex];
  }

  // When includeSelling is false, skip selling costs and taxes
  if (!includeSelling) {
    return {
      salePrice,
      totalSellingCosts: 0,
      loanPayoff,
      capitalGains: 0,
      taxOnGains: 0,
      netProceeds: salePrice - loanPayoff,
    };
  }

  const years = Math.floor(months / 12);
  const inflatedStagingCosts = inputs.stagingCosts * Math.pow(1 + inputs.inflationRate / 100, years);
  const agentFee = salePrice * (inputs.agentCommission / 100);
  const totalSellingCosts = agentFee + inflatedStagingCosts;

  const capitalGains = salePrice - inputs.purchasePrice - totalSellingCosts;

  const taxFreeLimitIndex = Math.max(0, Math.min(years - 1, inputs.taxFreeLimits.length - 1));
  const taxFreeLimit = inputs.taxFreeLimits[taxFreeLimitIndex] || 0;

  const taxableGains = Math.max(0, capitalGains - taxFreeLimit);
  const taxOnGains = taxableGains * (inputs.capitalGainsTax / 100);

  const netProceeds = salePrice - totalSellingCosts - loanPayoff - taxOnGains;

  return {
    salePrice,
    totalSellingCosts,
    loanPayoff,
    capitalGains,
    taxOnGains,
    netProceeds,
  };
}

export function calculateRentingNetWorth(
  inputs: CalculatorInputs,
  months: number,
  monthlyBuyingCosts: number[],
  monthlyRentingCosts: number[]
): number {
  const downpayment = inputs.purchasePrice - inputs.loanAmount;
  let investmentValue = downpayment - inputs.rentDeposit;
  const monthlyInvestmentRate = inputs.investmentReturnRate / 100 / 12;

  for (let i = 0; i < months; i++) {
    const monthlySavings = monthlyBuyingCosts[i] - monthlyRentingCosts[i];
    investmentValue += monthlySavings;
    investmentValue *= 1 + monthlyInvestmentRate;
  }

  const recoverableDeposit = inputs.rentDeposit * 0.75;
  return investmentValue + recoverableDeposit;
}

export function calculate(inputs: CalculatorInputs): CalculationResults {
  const costs = populateMonthlyCosts(inputs);
  const { effectiveLoanTerm, effectiveLoanAmount, refinanceCashOut } = getEffectiveLoanValues(inputs);
  const periods = getPeriods(effectiveLoanTerm, inputs.projectionYears);

  const keepTracking = calculateKeepInvestmentTracking(
    costs.monthlyBuyingCosts,
    inputs.investmentReturnRate,
    360,
    refinanceCashOut
  );

  // Calculate sale proceeds for all periods
  // For buy_vs_rent, respect the includeSelling flag; for sell_vs_keep, always include selling costs
  const shouldIncludeSelling = inputs.scenario === 'sell_vs_keep' ? true : inputs.includeSelling;
  const saleProceedsTable = periods.map((period) => {
    const proceeds = calculateSaleProceeds(inputs, period.months, costs.remainingLoanBalance, effectiveLoanAmount, shouldIncludeSelling);
    return {
      period: 'SALE ' + period.label,
      ...proceeds,
    };
  });

  const results: CalculationResults = {
    ...costs,
    ...keepTracking,
    periods,
    saleProceedsTable,
  };

  // Scenario-specific calculations
  if (inputs.scenario === 'buy_vs_rent') {
    // Amortization table
    if (inputs.loanAmount > 0) {
      const loanValues = getEffectiveLoanValues(inputs);
      const taxRate = inputs.mortgageInterestDeduction / 100;
      results.amortizationTable = periods.map((period) => {
        if (period.months === 0) {
          return {
            period: 'LOAN ' + period.label,
            principalPaid: 0,
            interestPaid: 0,
            taxDeduction: 0,
            effectiveInterest: 0,
            effectiveLoanPayment: 0,
            loanBalance: loanValues.effectiveLoanAmount,
          };
        }
        const monthIndex = Math.min(period.months - 1, costs.remainingLoanBalance.length - 1);
        const principalPaid = costs.cumulativePrincipalPaid[monthIndex];
        const interestPaid = costs.cumulativeInterestPaid[monthIndex];
        const taxDeduction = interestPaid * taxRate;
        const effectiveInterest = interestPaid - taxDeduction;
        return {
          period: 'LOAN ' + period.label,
          principalPaid,
          interestPaid,
          taxDeduction,
          effectiveInterest,
          effectiveLoanPayment: principalPaid + effectiveInterest,
          loanBalance: costs.remainingLoanBalance[monthIndex],
        };
      });
    }

    // Expenditure table - uses cumulative values from amortization table
    results.expenditureTable = periods.map((period, index) => {
      const downpayment = inputs.purchasePrice - inputs.loanAmount;
      let buyingExpenditure = downpayment;
      let rentingExpenditure = inputs.rentDeposit;

      for (let i = 0; i < period.months; i++) {
        buyingExpenditure += costs.monthlyBuyingCosts[i];
        rentingExpenditure += costs.monthlyRentingCosts[i];
      }

      // Get cumulative values directly from amortization table
      const amortRow = results.amortizationTable?.[index];
      const loanPayment = amortRow ? amortRow.principalPaid + amortRow.interestPaid : 0;
      const taxDeduction = amortRow?.taxDeduction ?? 0;
      const effectiveLoanPayment = amortRow?.effectiveLoanPayment ?? 0;

      // Calculate cumulative costs
      let cumulativeCosts = 0;
      for (let year = 0; year < period.months / 12; year++) {
        const inflationFactor = Math.pow(1 + inputs.inflationRate / 100, year);
        cumulativeCosts += (inputs.annualInsurance + inputs.annualTaxes - inputs.annualIncome) * inflationFactor;
      }

      return {
        period: 'EXP ' + period.label,
        loanPayment,
        taxDeduction,
        effectiveLoanPayment,
        costs: cumulativeCosts,
        buyingExpenditure,
        rentingExpenditure,
        difference: buyingExpenditure - rentingExpenditure,
      };
    });

    // Comparison table
    const loanValuesForComparison = getEffectiveLoanValues(inputs);
    results.comparisonTable = periods.map((period) => {
      const assetValue = calculateAssetValue(
        inputs.purchasePrice,
        period.months,
        inputs.appreciationRate
      );

      const downpayment = inputs.purchasePrice - inputs.loanAmount;
      const loanBalance = period.months === 0
        ? loanValuesForComparison.effectiveLoanAmount
        : costs.remainingLoanBalance[Math.min(period.months - 1, costs.remainingLoanBalance.length - 1)];

      let buyingNetWorth: number;
      if (inputs.includeSelling) {
        const { netProceeds } = calculateSaleProceeds(
          inputs,
          period.months,
          costs.remainingLoanBalance,
          effectiveLoanAmount
        );
        buyingNetWorth = netProceeds;
      } else {
        buyingNetWorth = assetValue - loanBalance;
      }

      const rentingNetWorth = calculateRentingNetWorth(
        inputs,
        period.months,
        costs.monthlyBuyingCosts,
        costs.monthlyRentingCosts
      );

      let cumulativeSavings = downpayment - inputs.rentDeposit;
      for (let i = 0; i < period.months; i++) {
        cumulativeSavings += costs.monthlyBuyingCosts[i] - costs.monthlyRentingCosts[i];
      }

      const recoverableDeposit = inputs.rentDeposit * 0.75;
      const marketReturn = rentingNetWorth - cumulativeSavings - recoverableDeposit;

      return {
        period: 'NET ' + period.label,
        assetValue,
        buyingNetWorth,
        cumulativeSavings,
        marketReturn,
        rentingNetWorth,
        difference: rentingNetWorth - buyingNetWorth,
      };
    });
  } else if (inputs.scenario === 'sell_vs_keep') {
    // sell_vs_keep scenario

    // Keep Expenses Breakdown table
    const loanValues = getEffectiveLoanValues(inputs);
    const taxRate = inputs.mortgageInterestDeduction / 100;

    // Amortization table for sell_vs_keep (if loan exists)
    if (inputs.loanAmount > 0) {
      results.amortizationTable = periods.map((period) => {
        if (period.months === 0) {
          return {
            period: 'LOAN ' + period.label,
            principalPaid: 0,
            interestPaid: 0,
            taxDeduction: 0,
            effectiveInterest: 0,
            effectiveLoanPayment: 0,
            loanBalance: loanValues.effectiveLoanAmount,
          };
        }
        const monthIndex = Math.min(period.months - 1, costs.remainingLoanBalance.length - 1);
        const principalPaid = costs.cumulativePrincipalPaid[monthIndex];
        const interestPaid = costs.cumulativeInterestPaid[monthIndex];
        const taxDeduction = interestPaid * taxRate;
        const effectiveInterest = interestPaid - taxDeduction;
        return {
          period: 'LOAN ' + period.label,
          principalPaid,
          interestPaid,
          taxDeduction,
          effectiveInterest,
          effectiveLoanPayment: principalPaid + effectiveInterest,
          loanBalance: costs.remainingLoanBalance[monthIndex],
        };
      });
    }

    results.keepExpensesTable = periods.map((period, index) => {
      const monthIndex = Math.min(period.months - 1, costs.cumulativeInterestPaid.length - 1);

      if (period.months === 0) {
        return {
          period: 'KEEP ' + period.label,
          loanPayment: 0,
          taxDeduction: 0,
          effectiveLoanPayment: 0,  // Negative = outflow
          incomeMinusCosts: 0,      // Positive = net income, Negative = net costs
          cumulativeExp: 0,         // Negative = net outflow
          investmentReturns: 0,
          netPosition: refinanceCashOut,
        };
      }

      // Get cumulative values from amortization table for consistency
      const amortRow = results.amortizationTable?.[index];
      const loanPayment = amortRow ? amortRow.principalPaid + amortRow.interestPaid : 0;
      const taxDeduction = amortRow?.taxDeduction ?? 0;
      const effectiveLoanPayment = amortRow?.effectiveLoanPayment ?? 0;

      // Calculate cumulative income minus costs (positive = income exceeds costs)
      let cumulativeIncomeMinusCosts = 0;
      for (let year = 0; year < period.months / 12; year++) {
        const inflationFactor = Math.pow(1 + inputs.inflationRate / 100, year);
        cumulativeIncomeMinusCosts += (inputs.annualIncome - inputs.annualInsurance - inputs.annualTaxes) * inflationFactor;
      }

      // Cumulative expenses up to this period (as negative = outflow)
      let cumulativeExp = 0;
      for (let i = 0; i < period.months; i++) {
        cumulativeExp += costs.monthlyBuyingCosts[i];
      }

      return {
        period: 'KEEP ' + period.label,
        loanPayment,
        taxDeduction,
        effectiveLoanPayment: -effectiveLoanPayment,  // Negative = outflow
        incomeMinusCosts: cumulativeIncomeMinusCosts, // Positive = net income, Negative = net costs
        cumulativeExp: -cumulativeExp,                // Negative = net outflow
        investmentReturns: keepTracking.monthlyKeepInvestmentReturns[monthIndex],
        netPosition: keepTracking.monthlyKeepNetPosition[monthIndex],
      };
    });

    results.sellVsKeepTable = periods.map((period) => {
      const keepNetPosition = period.months === 0
        ? 0
        : keepTracking.monthlyKeepNetPosition[Math.min(period.months - 1, keepTracking.monthlyKeepNetPosition.length - 1)];

      const { netProceeds } = calculateSaleProceeds(
        inputs,
        period.months,
        costs.remainingLoanBalance,
        effectiveLoanAmount
      );
      const keepNetWorth = netProceeds + keepNetPosition;

      // Calculate sell net worth (selling now and investing)
      const currentMarketValue = inputs.currentMarketValue || inputs.purchasePrice;
      const salePrice = currentMarketValue;
      const agentFee = salePrice * (inputs.agentCommission / 100);
      const totalSellingCosts = agentFee + inputs.stagingCosts;
      const loanPayoff = loanValues.effectiveLoanAmount;
      const capitalGains = salePrice - inputs.purchasePrice - totalSellingCosts;
      const taxFreeLimit = inputs.taxFreeLimits[0] || 0;
      const taxableGains = Math.max(0, capitalGains - taxFreeLimit);
      const taxOnGains = taxableGains * (inputs.capitalGainsTax / 100);
      const netProceedsNow = salePrice - totalSellingCosts - loanPayoff - taxOnGains;

      let sellNetWorth: number;
      let sellCumulativeExpenses: number | undefined;

      if (inputs.includeRentingSell) {
        let investmentValue = netProceedsNow - inputs.rentDeposit;
        const monthlyInvestmentRate = inputs.investmentReturnRate / 100 / 12;

        for (let i = 0; i < period.months; i++) {
          investmentValue -= costs.monthlyRentingCosts[i];
          investmentValue *= 1 + monthlyInvestmentRate;
        }

        const recoverableDeposit = inputs.rentDeposit * 0.75;
        sellNetWorth = investmentValue + recoverableDeposit;

        // Calculate cumulative rental expenses
        let cumulativeRentExpenses = inputs.rentDeposit;
        for (let i = 0; i < period.months; i++) {
          cumulativeRentExpenses += costs.monthlyRentingCosts[i];
        }
        cumulativeRentExpenses -= recoverableDeposit;
        sellCumulativeExpenses = cumulativeRentExpenses;
      } else {
        let investmentValue = netProceedsNow;
        const monthlyInvestmentRate = inputs.investmentReturnRate / 100 / 12;

        for (let i = 0; i < period.months; i++) {
          investmentValue *= 1 + monthlyInvestmentRate;
        }

        sellNetWorth = investmentValue;
      }

      return {
        period: 'NET ' + period.label,
        sellCumulativeExpenses,
        sellNetWorth,
        keepSaleProceeds: netProceeds,
        keepNetPosition,
        keepNetWorth: keepNetWorth,
        difference: keepNetWorth - sellNetWorth,
      };
    });
  } else if (inputs.scenario === 'payoff_vs_invest') {
    // payoff_vs_invest scenario (no mortgage interest deduction)
    const extraPayment = inputs.extraMonthlyPayment || 0;
    const extraUpfront = inputs.extraUpfrontPayment || 0;
    const monthlyRate = inputs.loanRate / 100 / 12;
    const monthlyInvestmentRate = inputs.investmentReturnRate / 100 / 12;
    const loanValues = getEffectiveLoanValues(inputs);
    const regularPayment = loanValues.monthlyLoanPayment;
    const loanTermMonths = inputs.remainingLoanTerm || inputs.loanTerm;

    // Calculate PAYOFF path monthly payment (may be recalculated if toggle is on)
    const upfrontPrincipal = Math.min(extraUpfront, loanValues.effectiveLoanAmount);
    const payoffStartingBalance = loanValues.effectiveLoanAmount - upfrontPrincipal;
    let payoffRegularPayment = regularPayment;
    if (inputs.recalculatePayment && extraUpfront > 0 && payoffStartingBalance > 0) {
      // Recalculate monthly payment based on reduced principal, same term
      payoffRegularPayment = calculateMonthlyPayment(payoffStartingBalance, monthlyRate, loanTermMonths);
    }

    // First calculate INVEST path to get effective payments for comparison
    const maxMonths = 360;
    const investLoanBalances: number[] = [...costs.remainingLoanBalance];
    const investInvestmentValues: number[] = new Array(maxMonths);
    const investCumulativeEffectivePayment: number[] = new Array(maxMonths);
    const investCumulativeContributions: number[] = new Array(maxMonths);
    const investMonthlyEffectivePayment: number[] = new Array(maxMonths);

    // INVEST path: invest the upfront amount instead of paying down loan
    let investInvestment = extraUpfront;
    let investBalance = loanValues.effectiveLoanAmount;
    let investTotalEffectivePayment = 0;
    let investTotalContributions = extraUpfront;

    for (let i = 0; i < maxMonths; i++) {
      if (i < loanTermMonths && investBalance > 0) {
        // Still paying loan - invest extra only
        const interestPayment = investBalance * monthlyRate;

        investInvestment += extraPayment;
        investTotalEffectivePayment += regularPayment;
        investTotalContributions += extraPayment;
        investMonthlyEffectivePayment[i] = regularPayment;

        // Update loan balance
        const principalPayment = regularPayment - interestPayment;
        investBalance = Math.max(0, investBalance - principalPayment);
      } else {
        // Loan paid off - invest extra only
        investInvestment += extraPayment;
        investTotalContributions += extraPayment;
        investMonthlyEffectivePayment[i] = 0;
      }

      // Compound investment
      investInvestment *= 1 + monthlyInvestmentRate;
      investInvestmentValues[i] = investInvestment;
      investCumulativeEffectivePayment[i] = investTotalEffectivePayment;
      investCumulativeContributions[i] = investTotalContributions;
    }

    // Now calculate PAYOFF path with deficit tracking
    const payoffLoanBalances: number[] = new Array(maxMonths);
    const payoffInvestmentValues: number[] = new Array(maxMonths);
    const payoffCumulativePrincipal: number[] = new Array(maxMonths);
    const payoffCumulativeInterest: number[] = new Array(maxMonths);
    const payoffCumulativeEffectivePayment: number[] = new Array(maxMonths);
    const payoffCumulativeContributions: number[] = new Array(maxMonths);
    const payoffCumulativeReturns: number[] = new Array(maxMonths);

    // PAYOFF path: apply upfront payment to reduce starting loan balance
    // Note: upfront payment is same cost to both paths, just allocated differently
    // So payoffNetPosition starts at 0 - deficit only tracks extra monthly payment difference
    let payoffBalance = payoffStartingBalance;
    let payoffNetPosition = 0;
    let loanPaidOff = payoffBalance <= 0;
    let payoffTotalPrincipal = upfrontPrincipal;
    let payoffTotalInterest = 0;
    let payoffTotalEffectivePayment = extraUpfront; // Upfront counts as effective payment
    let payoffTotalContributions = 0;
    let payoffTotalReturns = 0;

    for (let i = 0; i < maxMonths; i++) {
      if (!loanPaidOff && payoffBalance > 0) {
        // Still paying off loan with extra payments - no investment yet
        const interestPayment = payoffBalance * monthlyRate;
        const principalFromPayment = payoffRegularPayment - interestPayment;
        const desiredPrincipalPayment = principalFromPayment + extraPayment;
        // Cap principal payment at remaining balance (final month adjustment)
        const actualPrincipalPayment = Math.min(desiredPrincipalPayment, payoffBalance);
        // Payment = interest + actual principal (no tax deduction)
        const payoffPayment = interestPayment + actualPrincipalPayment;

        payoffTotalInterest += interestPayment;
        payoffTotalPrincipal += actualPrincipalPayment;
        payoffTotalEffectivePayment += payoffPayment;
        payoffBalance = payoffBalance - actualPrincipalPayment;

        if (payoffBalance <= 0) {
          loanPaidOff = true;
          payoffBalance = 0;
        }
        // payoffNetPosition stays 0 while paying off loan
      } else {
        // Loan paid off - contribute freed-up payments to investment
        const contribution = payoffRegularPayment + extraPayment;
        payoffNetPosition += contribution;
        payoffTotalContributions += contribution;

        // Compound investment returns
        const monthlyReturn = payoffNetPosition * monthlyInvestmentRate;
        payoffTotalReturns += monthlyReturn;
        payoffNetPosition += monthlyReturn;
      }

      payoffLoanBalances[i] = payoffBalance;
      payoffInvestmentValues[i] = payoffNetPosition;
      payoffCumulativePrincipal[i] = payoffTotalPrincipal;
      payoffCumulativeInterest[i] = payoffTotalInterest;
      payoffCumulativeEffectivePayment[i] = payoffTotalEffectivePayment;
      payoffCumulativeContributions[i] = payoffTotalContributions;
      payoffCumulativeReturns[i] = payoffTotalReturns;
    }

    // Build payoff amortization table
    results.payoffAmortizationTable = periods.map((period) => {
      if (period.months === 0) {
        return {
          period: 'PAYOFF ' + period.label,
          principalPaid: upfrontPrincipal,
          interestPaid: 0,
          taxDeduction: 0,
          effectiveInterest: 0,
          effectiveLoanPayment: extraUpfront,
          loanBalance: payoffStartingBalance,
        };
      }
      const monthIndex = Math.min(period.months - 1, maxMonths - 1);
      const principalPaid = payoffCumulativePrincipal[monthIndex];
      const interestPaid = payoffCumulativeInterest[monthIndex];
      return {
        period: 'PAYOFF ' + period.label,
        principalPaid,
        interestPaid,
        taxDeduction: 0,
        effectiveInterest: interestPaid,
        effectiveLoanPayment: payoffCumulativeEffectivePayment[monthIndex],
        loanBalance: payoffLoanBalances[monthIndex],
      };
    });

    // Build invest amortization table (regular payments)
    results.investAmortizationTable = periods.map((period) => {
      if (period.months === 0) {
        return {
          period: 'INVEST ' + period.label,
          principalPaid: 0,
          interestPaid: 0,
          taxDeduction: 0,
          effectiveInterest: 0,
          effectiveLoanPayment: 0,
          loanBalance: loanValues.effectiveLoanAmount,
        };
      }
      const monthIndex = Math.min(period.months - 1, maxMonths - 1);
      const principalPaid = costs.cumulativePrincipalPaid[monthIndex];
      const interestPaid = costs.cumulativeInterestPaid[monthIndex];
      return {
        period: 'INVEST ' + period.label,
        principalPaid,
        interestPaid,
        taxDeduction: 0,
        effectiveInterest: interestPaid,
        effectiveLoanPayment: investCumulativeEffectivePayment[monthIndex],
        loanBalance: investLoanBalances[monthIndex],
      };
    });

    // Build comparison table
    results.payoffVsInvestTable = periods.map((period) => {
      if (period.months === 0) {
        return {
          period: 'NET ' + period.label,
          payoffLoanBalance: loanValues.effectiveLoanAmount,
          payoffInvestmentValue: 0,
          payoffWealth: -loanValues.effectiveLoanAmount,
          investLoanBalance: loanValues.effectiveLoanAmount,
          investInvestmentValue: 0,
          investWealth: -loanValues.effectiveLoanAmount,
          difference: 0,
        };
      }

      const monthIndex = Math.min(period.months - 1, maxMonths - 1);

      const payoffLoanBal = payoffLoanBalances[monthIndex];
      const payoffInvest = payoffInvestmentValues[monthIndex];
      const payoffWealth = payoffInvest - payoffLoanBal;

      const investLoanBal = investLoanBalances[monthIndex];
      const investInvest = investInvestmentValues[monthIndex];
      const investWealth = investInvest - investLoanBal;

      return {
        period: 'NET ' + period.label,
        payoffLoanBalance: payoffLoanBal,
        payoffInvestmentValue: payoffInvest,
        payoffWealth,
        investLoanBalance: investLoanBal,
        investInvestmentValue: investInvest,
        investWealth,
        difference: payoffWealth - investWealth,
      };
    });
  }

  return results;
}
