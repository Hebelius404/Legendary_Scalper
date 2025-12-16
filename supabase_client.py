
import os
import time
import json
import logging
from typing import Dict, List, Optional
from datetime import datetime
from supabase import create_client, Client
import threading
from dotenv import load_dotenv

# Load env variables (force reload in case they changed)
load_dotenv()

logger = logging.getLogger("SupabaseClient")

class SupabaseManager:
    """
    Manages connection to Supabase for remote control/monitoring
    """
    def __init__(self):
        # Try env first, then fallback to provided keys
        self.url = os.getenv("SUPABASE_URL") or "https://rzwjvmtypqmnrjcxaymy.supabase.co"
        self.key = os.getenv("SUPABASE_KEY") or "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InJ6d2p2bXR5cHFtbnJqY3hheW15Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjU0MDA0MDYsImV4cCI6MjA4MDk3NjQwNn0.FtQNBLhk43cG3CrbkaQ2rmhM60xeMKuvNumaA8mh5xg"
        
        logger.info(f"DEBUG: Supabase URL: {self.url}")
        logger.info(f"DEBUG: Supabase Key len: {len(self.key) if self.key else 0}")

        self.client: Optional[Client] = None
        self.last_sync = 0
        self.connected = False
        self.bot_id = 1  # Singleton ID for this bot
        
        if self.url and self.key:
            try:
                self.client = create_client(self.url, self.key)
                self.connected = True
                logger.info("☁️ Supabase client initialized")
            except Exception as e:
                logger.error(f"❌ Failed to init Supabase: {e}")
        else:
            logger.warning("⚠️ No Supabase credentials found in .env")

    def sync_state(self, bot_status: Dict):
        """
        Push current bot state to Supabase 'bot_state' table
        """
        if not self.connected or not self.client:
            return

        # Throttle sync to once per second
        if time.time() - self.last_sync < 1.0:
            return

        try:
            # Prepare data
            # HACK: Wrapping config into positions_json to avoid schema migration
            # { "positions": {...}, "config": {...} }
            positions_wrapper = {
                "positions": bot_status.get("positions", {}),
                "config": bot_status.get("config", {}),
                # Pack stats here to avoid schema migration
                "balance": bot_status.get("balance", 0),
                "total_realized_pnl": bot_status.get("total_realized_pnl", 0),
                "wins": bot_status.get("wins", 0),
                "losses": bot_status.get("losses", 0),
                "break_evens": bot_status.get("break_evens", 0),
                "gross_profit": bot_status.get("gross_profit", 0),
                "gross_loss": bot_status.get("gross_loss", 0)
            }

            data = {
                "id": self.bot_id,
                "is_running": bot_status.get("is_running", False),
                "last_heartbeat": datetime.now().isoformat(),
                "balance": bot_status.get("balance", 0),
                "active_positions_count": bot_status.get("active_positions", 0),
                "total_pnl": bot_status.get("total_unrealized_pnl", 0),
                "positions_json": json.dumps(positions_wrapper),
                "updated_at": datetime.now().isoformat()
            }
            
            # Upsert (Insert or Update)
            self.client.table("bot_state").upsert(data).execute()
            self.last_sync = time.time()
            
        except BlockingIOError:
            pass # Ignore benign socket wait errors (WinError 10035)
        except Exception as e:
            # Check for Windows socket error 10035 or connection reset
            error_str = str(e)
            if "10035" in error_str or "ConnectionTerminated" in error_str:
                pass
            else:
                logger.warning(f"☁️ Sync failed (retrying): {e}")
            # self.connected = False  <-- DISABLED: Don't kill connection permantly on error
            
    def poll_commands(self) -> List[Dict]:
        """
        Check 'commands' table for new PENDING commands
        """
        if not self.connected or not self.client:
            return []

        try:
            response = self.client.table("commands")\
                .select("*")\
                .eq("status", "PENDING")\
                .execute()
            
            commands = response.data
            
            # Mark them as EXECUTED so we don't run them twice
            for cmd in commands:
                self.client.table("commands")\
                    .update({"status": "EXECUTED", "executed_at": datetime.now().isoformat()})\
                    .eq("id", cmd['id'])\
                    .execute()
                    
            return commands
            
        except Exception as e:
            # Suppress noisy connection errors
            error_str = str(e)
            if "ConnectionTerminated" in error_str:
                logger.warning(f"☁️ Connection flake (retrying)...")
            else:
                logger.warning(f"☁️ Command poll failed: {e}")
            return []

    def log(self, level: str, message: str):
        """
        Push log entry to 'logs' table
        """
        if not self.connected or not self.client:
            return
            
        try:
            self.client.table("logs").insert({
                "level": level,
                "message": message,
                "timestamp": datetime.now().isoformat()
            }).execute()
        except Exception as e:
            pass # Fail silently for logs

# Singleton instance
supabase_manager = SupabaseManager()
