# Changelog

## [v1.7] (Safety Update) - 2025-12-16
### New Features
- **RSI Circuit Breaker:**
    - Prevents adding Martingale steps if RSI > 90 (configurable via `MARTINGALE_RSI_MAX_LIMIT`).
    - Stops the bot from "fighting the trend" during parabolic pumps ("God Candles").
    - Displays `🚫 RSI Circuit Breaker` warning in logs.
- **Dynamic Step Spacing (Volatility Adaptive):**
    - Automatically widens the distance between steps during high volatility.
    - Uses a `Volatility Multiplier` (1.0x - 3.0x) based on the ratio of the current candle size to the average range.
    - Reduces margin exposure during rapid price moves by requiring deeper pullbacks for step entries.

### Fixes
- **Dashboard:** Updated version tag to v1.7.

## [v1.6.1] (Balance Fix) - 2025-12-16
### Fixes
- **Balance Calculation:** Update balance display to show Wallet Balance (Total Equity) instead of Available Balance.    

## [v1.6] (Detail+) - 2025-12-15
### Changes
- **Detailed Stats:** Added Win Rate, Break-Even count, and Profit Factor cards.
- **Fixed Stats Display:** Solved bug where stats were showing $0.00 due to database schema.
- **Dashboard Refactor:** Improved grid layout for better readability.

## [v1.5] (Charts Fixed) - 2025-12-15
### Changes
- Fixed TradingView Chart symbol format (added `.P` suffix for Perpetual Contracts).
- Increased Chart container height to 500px.

## [v1.4] (Bugfix Update) - 2025-12-14
### Changes
- Fixed `get_position` crash when position was missing.
- Added `get_usdt_balance` method to BinanceClient.
- Improved logging for position scanning.

## [v1.3] (Supabase Integration) - 2025-12-12
### Changes
- Added meaningful status messages for "Waiting..." state.
- Integrated `supabase_client.py` for remote command handling (Start/Stop/Liquidate).
- Added `sync_state` to push bot status to Supabase DB.

## [v1.2] (Visual Orders) - 2025-12-10
### Changes
- Added visual orders on chart for TP/SL/Steps.
- Added `place_visual_orders` logic.

## [v1.1] (Initial Release) - 2025-12-01
- Core Martingale Strategy.
- Binance Futures API integration.
