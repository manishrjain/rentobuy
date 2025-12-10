<script lang="ts">
  import { getMarketYears, getMarketAverages, getLastUpdated } from '../lib/marketData';

  const years = getMarketYears();
  const avg5y = getMarketAverages(5);
  const avg10y = getMarketAverages(10);
  const avg15y = getMarketAverages(15);
  const lastUpdated = getLastUpdated();

  function formatPercent(value: number): string {
    return `${value.toFixed(2)}%`;
  }
</script>

<section class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg font-mono">
  <h2 class="section-title">Market Data</h2>
  <div class="table-container">
    <table class="data-table">
      <thead>
        <tr>
          <th>Period</th>
          <th class="text-right">Inflation</th>
          <th class="text-right">VOO</th>
          <th class="text-right">QQQ</th>
          <th class="text-right">VTI</th>
          <th class="text-right">BND</th>
          <th class="text-right">60/40 VTI/BND</th>
        </tr>
      </thead>
      <tbody>
        {#each years as row}
          <tr>
            <td class="font-mono">MRKT {row.year}</td>
            <td class="text-right font-mono">{row.inflation !== null ? formatPercent(row.inflation) : '—'}</td>
            <td class="text-right font-mono">{formatPercent(row.voo)}</td>
            <td class="text-right font-mono">{formatPercent(row.qqq)}</td>
            <td class="text-right font-mono">{formatPercent(row.vti)}</td>
            <td class="text-right font-mono">{formatPercent(row.bnd)}</td>
            <td class="text-right font-mono">{formatPercent(row.mix6040)}</td>
          </tr>
        {/each}
        <tr class="avg-row">
          <td class="font-mono text-light-cyan dark:text-monokai-cyan">MRKT Avg (5y)</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg5y.inflation)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg5y.voo)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg5y.qqq)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg5y.vti)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg5y.bnd)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg5y.mix6040)}</td>
        </tr>
        <tr class="avg-row">
          <td class="font-mono text-light-cyan dark:text-monokai-cyan">MRKT Avg (10y)</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg10y.inflation)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg10y.voo)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg10y.qqq)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg10y.vti)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg10y.bnd)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg10y.mix6040)}</td>
        </tr>
        <tr class="avg-row">
          <td class="font-mono text-light-cyan dark:text-monokai-cyan">MRKT Avg (15y)</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg15y.inflation)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg15y.voo)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg15y.qqq)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg15y.vti)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg15y.bnd)}</td>
          <td class="text-right font-mono text-light-cyan dark:text-monokai-cyan">{formatPercent(avg15y.mix6040)}</td>
        </tr>
      </tbody>
    </table>
  </div>
  <p class="help-text mt-2">
    Historical annual returns for major ETFs. VOO = S&P 500, QQQ = Nasdaq 100, VTI = Total Stock Market, BND = Total Bond Market.
    Last updated: {lastUpdated}
  </p>
</section>

<style>
  .avg-row {
    @apply border-t-2 border-light-border dark:border-monokai-border;
  }
</style>
