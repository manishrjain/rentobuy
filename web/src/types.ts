export type ScenarioType = 'buy_vs_rent' | 'sell_vs_keep';

export interface CalculatorInputs {
  // Scenario selection
  scenario: ScenarioType;

  // Economic assumptions
  inflationRate: number;
  investmentReturnRate: number;
  include30Year: boolean;

  // Buying/Asset
  purchasePrice: number;
  currentMarketValue?: number; // Only for sell_vs_keep
  loanAmount: number;
  loanRate: number;
  loanTerm: number; // in months
  remainingLoanTerm?: number; // Only for sell_vs_keep, in months
  annualInsurance: number;
  annualTaxes: number;
  monthlyExpenses: number;
  appreciationRate: number[]; // Array for different years

  // Renting
  rentDeposit: number;
  monthlyRent: number;
  annualRentCosts: number;
  otherAnnualCosts: number;

  // Selling
  includeSelling: boolean;
  includeRentingSell?: boolean; // Only for sell_vs_keep
  agentCommission: number;
  stagingCosts: number;
  taxFreeLimits: number[]; // Array for different years
  capitalGainsTax: number;
}

export interface Period {
  label: string;
  months: number;
}

export interface AmortizationRow {
  period: string;
  principalPaid: number;
  interestPaid: number;
  loanBalance: number;
}

export interface ExpenditureRow {
  period: string;
  buyingExpenditure: number;
  rentingExpenditure: number;
  difference: number;
}

export interface SaleProceedsRow {
  period: string;
  salePrice: number;
  totalSellingCosts: number;
  loanPayoff: number;
  capitalGains: number;
  taxOnGains: number;
  netProceeds: number;
}

export interface ComparisonRow {
  period: string;
  assetValue: number;
  buyingNetWorth: number;
  cumulativeSavings: number;
  marketReturn: number;
  rentingNetWorth: number;
  difference: number;
}

export interface SellVsKeepRow {
  period: string;
  sellCumulativeExpenses?: number;
  sellNetWorth: number;
  keepNetPosition: number;
  keepNetProceeds: number;
  difference: number;
}

export interface KeepExpensesRow {
  period: string;
  loanPayment: number;
  taxInsurance: number;
  otherCosts: number;
  cumulativeExp: number;
  investmentVal: number;
  netPosition: number;
}

export interface CalculationResults {
  // Common arrays for calculations
  monthlyBuyingCosts: number[];
  monthlyRentingCosts: number[];
  remainingLoanBalance: number[];
  cumulativePrincipalPaid: number[];
  cumulativeInterestPaid: number[];

  // For KEEP scenario
  monthlyKeepInvestmentValue: number[];
  monthlyKeepRealCosts: number[];
  monthlyKeepNetPosition: number[];

  // Table data
  periods: Period[];
  amortizationTable?: AmortizationRow[];
  expenditureTable?: ExpenditureRow[];
  saleProceedsTable: SaleProceedsRow[];
  comparisonTable?: ComparisonRow[];
  sellVsKeepTable?: SellVsKeepRow[];
  keepExpensesTable?: KeepExpensesRow[];
}
