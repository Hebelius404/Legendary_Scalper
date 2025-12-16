"""
Legendary Martingale Scalper - Counter-Trend SHORT Strategy

╔══════════════════════════════════════════════════════════════╗
║  🎰 LEGENDARY MARTINGALE SCALPER 🎰                          ║
║                                                              ║
║  Strategy: Counter-Trend Martingale SHORT                   ║
║  Target: Highly Pumped Coins (20%+)                         ║
║                                                              ║
║  "The bot hunts like an eagle"                              ║
╚══════════════════════════════════════════════════════════════╝

Finds pumped coins (20%+) and shorts on exhaustion signals
"""

import sys
import time
import threading # Added for SupabaseLogHandler
import signal as sig
import logging # Added for SupabaseLogHandler
from datetime import datetime, timezone

import config
from binance_client import BinanceClient
from order_executor import OrderExecutor
from pump_detector import PumpDetector
from martingale_manager import MartingaleManager
from position_watcher import PositionWatcher
from grok_client import GrokClient
from logger import logger
from supabase_client import supabase_manager # Remote Control

class SupabaseLogHandler(logging.Handler):
    """Sends logs to Supabase"""
    def emit(self, record):
        try:
            msg = self.format(record)
            # Run in thread to avoid blocking main loop
            threading.Thread(target=supabase_manager.log, args=(record.levelname, msg)).start()
        except:
            pass

# Attach Handler
supabase_handler = SupabaseLogHandler()
supabase_handler.setFormatter(logging.Formatter('%(asctime)s | %(levelname)s | %(message)s', datefmt='%H:%M:%S'))
logger.addHandler(supabase_handler)


class LegendaryScalper:
    """
    Counter-trend Martingale scalper
    
    Strategy:
    1. Find coins pumped >30% in 24h
    2. Wait for exhaustion signals (RSI overbought, volume decline)
    3. Enter SHORT with small margin
    4. Add steps as price goes higher (averaging up)
    5. Close half when price returns to average
    6. Take full profit on reversal
    """
    
    def __init__(self):
        self.running = False
        
        logger.info("🎰 Initializing Legendary Scalper...")
        
        # Core components
        self.client = BinanceClient()
        self.executor = OrderExecutor(self.client)
        
        # Martingale components
        self.pump_detector = PumpDetector(self.client)
        self.martingale = MartingaleManager(self.client, self.executor)
        self.watcher = PositionWatcher(self.client, self.martingale, self.pump_detector)
        
        # Grok AI for sentiment analysis
        self.grok = GrokClient() if getattr(config, 'GROK_ENABLED', True) else None
        
        # Tracking
        self.scan_count = 0
        self.last_pump_scan = 0
        self.pump_scan_interval = 90  # Scan for pumps every 90 seconds
        self.position_sync_interval = 30  # Sync with Binance every 30 cycles (~5 min)
        
        logger.info("✅ Legendary Scalper initialized!")
        self._print_config()
    
    def _print_config(self):
        """Print configuration summary"""
        logger.info("📊 Configuration:")
        logger.info(f"   Min Pump: {config.MARTINGALE_MIN_PUMP}%")
        logger.info(f"   Max Positions: {config.MARTINGALE_MAX_POSITIONS}")
        logger.info(f"   Steps: {config.MARTINGALE_STEPS}")
        logger.info(f"   Total Max Margin: ${sum(config.MARTINGALE_STEPS)}")
    
    def startup_checks(self) -> bool:
        """Perform startup checks"""
        logger.info("Running startup checks...")
        
        try:
            # Test API connection
            server_time = self.client.get_server_time()
            logger.info(f"✅ API Connected (Server time: {server_time})")
            
            # Check balance
            balance = self.client.get_usdt_balance()
            logger.info(f"✅ USDT Balance: {balance:.2f}")
            
            if balance < sum(config.MARTINGALE_STEPS):
                logger.warning(f"⚠️ Balance may be insufficient for full Martingale")
            
            # Recover existing positions from Binance
            logger.info("♻️ Checking for existing positions...")
            recovered = self.martingale.recover_positions()
            if recovered > 0:
                logger.info(f"✅ Recovered {recovered} positions - will continue managing them!")
            
            # Sync to ensure internal state matches Binance
            self.martingale.sync_positions()
            

            
            # Initial pump scan
            logger.info("🔍 Scanning for pumped coins...")
            pumped = self.pump_detector.find_pumped_coins()
            logger.info(f"✅ Found {len(pumped)} coins with >{config.MARTINGALE_MIN_PUMP}% pump")
            
            return True
            
        except Exception as e:
            logger.error(f"Startup check failed: {e}")
            return False
    
    def run_cycle(self):
        """Run a single scan cycle"""
        self.scan_count += 1
        current_time = time.time()
        
        # 1. Check existing positions (always)
        logger.info(f"📊 Cycle #{self.scan_count} | Checking positions...")
        actions = self.watcher.check_positions()
        
        if actions['steps_added']:
            for sym in actions['steps_added']:
                logger.info(f"🎰 Added step for {sym}")
        
        if actions['half_closed']:
            for sym in actions['half_closed']:
                logger.info(f"✂️ Half-closed {sym}")
        
        if actions['closed']:
            for sym in actions['closed']:
                logger.info(f"💰 Closed {sym}")
        
        if actions['emergency_closed']:
            for sym in actions['emergency_closed']:
                logger.warning(f"🚨 Emergency closed {sym}")
        
        # 2. Scan for new opportunities (every pump_scan_interval)
        if current_time - self.last_pump_scan >= self.pump_scan_interval:
            self.last_pump_scan = current_time
            
            logger.info("🔍 Scanning for pumped coins...")
            opportunities = self.watcher.scan_for_new_entries()
            
            for opp in opportunities[:5]:  # Open up to 5 positions per scan
                if self.martingale.can_open_new_position():
                    symbol = opp['symbol']
                    pump = opp['pump']
                    
                    # Check 1h trend for multi-timeframe confirmation
                    trend_check = self.pump_detector.check_1h_trend(symbol)
                    if not trend_check.get('ok_to_short', True):
                        logger.info(f"⏭️ Skipping {symbol} - {trend_check.get('reason')}")
                        continue
                    logger.info(f"📊 1h Check: {symbol} - {trend_check.get('reason')}")
                    
                    # Check RSI - only enter when overbought (RSI > 70)
                    rsi = self.client.calculate_rsi(symbol)
                    min_rsi = getattr(config, 'MARTINGALE_MIN_RSI', 70)
                    if rsi > 0 and rsi < min_rsi:
                        logger.info(f"⏭️ Skipping {symbol} - RSI {rsi:.1f} < {min_rsi} (not overbought)")
                        continue
                    if rsi > 0:
                        logger.info(f"📊 RSI Check: {symbol} RSI={rsi:.1f} ✅ (>{min_rsi})")
                    
                    # Check Grok sentiment for high pumps
                    if self.grok and pump >= 40:
                        sentiment = self.grok.is_good_short_entry(symbol, pump)
                        if not sentiment.get('is_good', True):
                            logger.info(f"⏭️ Skipping {symbol} - Grok: {sentiment.get('reason')}")
                            continue
                        logger.info(f"🤖 Grok: {symbol} FOMO {sentiment.get('fomo_level', 0)}%")
                    
                    logger.info(f"🎯 Opening SHORT: {symbol} (Pump: +{pump:.1f}%)")
                    self.martingale.open_position(symbol, opp['price'])
        
        # 3. Periodic sync with Binance (every N cycles)
        if self.scan_count % self.position_sync_interval == 0:
            self.martingale.sync_positions()
        
        # 4. Log status
        self.watcher.log_status()
        
        # 5. Sync to Supabase (Remote Control)
        # Get full status
        status = self.martingale.get_status()
        # Add extra info
        status['is_running'] = self.running
        status['balance'] = self.martingale.get_balance()
        
        # Push state
        supabase_manager.sync_state(status)
        
        # Check for remote commands
        commands = supabase_manager.poll_commands()
        for cmd in commands:
            command_type = cmd.get('command')
            if command_type == 'STOP':
                logger.info("📱 Received remote STOP command!")
                self.running = False
            elif command_type == 'START': # Already running, but good to ack
                logger.info("📱 Received remote START command (already running)")

    
    def process_remote_commands(self) -> bool:
        """
        Check for remote commands and update state. 
        Returns True if state changed (e.g. Start -> Stop).
        """
        changed = False
        try:
            commands = supabase_manager.poll_commands()
            for cmd in commands:
                command_type = cmd.get('command')
                if command_type == 'START' and not self.running:
                    logger.info("🚀 Received remote START command!")
                    self.running = True
                    changed = True
                elif command_type == 'STOP' and self.running:
                    logger.info("📱 Received remote STOP command!")
                    self.running = False
                    changed = True
        except Exception as e:
            logger.error(f"Command processing error: {e}")
            
        return changed

    def smart_sleep(self, duration: float):
        """
        Sleep for 'duration' seconds, but check for commands every 0.5s.
        Returns early if state changes.
        """
        interval = 0.5
        end_time = time.time() + duration
        
        while time.time() < end_time:
            # Poll for commands
            if self.process_remote_commands():
                # State changed (e.g. STOP received), exit sleep immediately
                return
            
            # Calculate sleep time
            remaining = end_time - time.time()
            if remaining <= 0:
                break
                
            time.sleep(min(remaining, interval))
    
    def _get_full_status(self):
        """Helper to build full status dict for Supabase"""
        status = self.martingale.get_status()
        status['is_running'] = self.running
        status['balance'] = self.martingale.get_balance()
        status['config'] = {
            'min_pump': config.MARTINGALE_MIN_PUMP,
            'max_margin': sum(config.MARTINGALE_STEPS),
            'max_positions': config.MAX_OPEN_POSITIONS,
            'steps': config.MARTINGALE_STEPS
        }
        return status

    def run(self):
        """
        Main Application Loop: Handles transitions between STANDBY and TRADING
        """
        if not self.startup_checks():
            logger.error("Startup failed!")
            return

        logger.info("📡 Bot is ready. Waiting for commands...")
        print("="*65)
        print("🤖 BOT IS ONLINE")
        print("📡 Status: STANDBY (Waiting for remote START)")
        print("="*65)

        try:
            while True:
                # 1. Initial Check (in case we loop back fast)
                if self.process_remote_commands():
                    # If state changed (START/STOP), sync IMMEDIATELY so dashboard updates fast
                    logger.info(f"🔄 State change detected. Syncing... (Running: {self.running})")
                    supabase_manager.sync_state(self._get_full_status())

                if self.running:
                    # --- TRADING MODE ---
                    self.run_cycle()
                    
                    # Log & Sync State
                    supabase_manager.sync_state(self._get_full_status())
                    
                    # Responsive Sleep (10s -> polls every 0.5s)
                    self.smart_sleep(config.SCAN_INTERVAL_SECONDS)
                    
                else:
                    # --- STANDBY MODE ---
                    # Sync Standby state WITH Config
                    supabase_manager.sync_state(self._get_full_status())
                    
                    # Responsive Sleep (Sync every 5s, Poll every 0.5s)
                    self.smart_sleep(5.0)

        except KeyboardInterrupt:
            logger.info("⛔ Hard Stop received (Ctrl+C)")
            self.stop()



    def stop(self):
        """Stop the bot logic (remains in standby unless hard exit)"""
        self.running = False
        logger.info("🛑 Bot trading logic stopped. Entering STANDBY mode.")
        
        # Log final status
        if hasattr(self, 'martingale'):
            status = self.martingale.get_status()
            logger.info(f"📊 Final Status: {status['active_positions']} open positions")

def signal_handler(signum, frame):
    """Handle Ctrl+C"""
    logger.info("Received stop signal... Exiting.")
    sys.exit(0)

if __name__ == "__main__":
    sig.signal(sig.SIGINT, signal_handler)
    
    bot = LegendaryScalper()
    bot.run()
