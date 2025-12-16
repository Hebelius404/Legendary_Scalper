# Changelog

All notable changes to the Legendary Scalper project will be documented in this file.

## [1.6.1] - 2025-12-14

### Added
- **Enhanced Log Visibility:** Added explicit terminal logs for the bot's scanning cycle.
    - `🔍 Scanning for pumped coins...` (Every 90s)
    - `⏸️ Scan skipped: Max positions reached` (Clear feedback when fully loaded)

### Fixed
- **Safety:** Enabled **Hard Stop** functionality. Previously, it only logged a warning. Now it will actively close positions that hit the `$100` USD loss or `35%` drawdown limit.
- **Dashboard:**
    - Fixed **Balance Display** (was showing $0.00).
    - Added **Total P&L (All Time)** card. This tracks your lifetime profits persistently in `total_pnl.json` (survives restarts).
    - Added **Win Rate** card with **Break-Even Detection**. Trades with PnL between `-$0.50` and `$0.50` are counted as Break-Even (e.g., `90% (9W/1L/2BE)`).
    - Added **Profit Factor** card (Gross Profit / Gross Loss) to track risk efficiency.
    - Updated Dashboard layout to a cleaner 3-row grid for stats.
    - **Fixed Missing Stats Bug:** Solved an issue where Balance, Win Rate, and PnL were showing as 0 due to database schema limits.
    - **Fixed Balance Calculation:** Now displaying **Wallet Balance** (Total Equity) instead of Available Balance (Free Margin), so it matches Binance UI exactly.
    - Fixed TradingView symbol format to use `.P` suffix.
    - Fixed TradingView symbol format to use `.P` suffix (e.g., `BTCUSDT.P`) ensuring Perpetual Futures data is displayed.
    - Increased Chart container height to **500px** for better readability.
- **Dashboard Steps Display:** Fixed "undefined" quantity error in the Steps History tab by correctly mapping `entry.quantity`.
- **Authentication:** Fixed infinite 401 loop by adding aggressively auto-logout and Supabase key sanitization on connection errors.
- **Version Tag:** Updated dashboard version display to `v1.6 (Detail+)`.

## [1.5.0] - Previous Release

### Added
- **Future Step Projections:** Dashboard now shows estimated prices and margins for future Martingale steps.
- **Detailed Positions:** New tabbed modal (Overview, Chart, Steps) for active positions.
