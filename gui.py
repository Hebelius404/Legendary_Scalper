
import customtkinter as ctk
import threading
import logging
import sys
import os
from datetime import datetime
from tkinter import messagebox
from typing import Dict, Optional

# Import the bot
try:
    from legendary_scalper import LegendaryScalper
    import config
except ImportError:
    messagebox.showerror("Error", "Could not import LegendaryScalper modules. Make sure you are in the correct directory.")
    sys.exit(1)

# Configure Appearance
ctk.set_appearance_mode("Dark")
ctk.set_default_color_theme("green")

class TextHandler(logging.Handler):
    """Custom logging handler to send logs to the text box"""
    def __init__(self, text_widget):
        super().__init__()
        self.text_widget = text_widget

    def emit(self, record):
        msg = self.format(record)
        def append():
            try:
                self.text_widget.configure(state="normal")
                self.text_widget.insert("end", msg + "\n")
                self.text_widget.see("end")
                self.text_widget.configure(state="disabled")
            except:
                pass
        self.text_widget.after(0, append)

class LegendaryGUI(ctk.CTk):
    def __init__(self):
        super().__init__()

        self.title("Legendary Martingale Scalper 🎰")
        self.geometry("1100x700")
        
        # Bot Instance
        self.bot: Optional[LegendaryScalper] = None
        self.bot_thread: Optional[threading.Thread] = None

        # Grid Layout
        self.grid_columnconfigure(1, weight=1)
        self.grid_rowconfigure(0, weight=1)

        # Sidebar
        self.sidebar_frame = ctk.CTkFrame(self, width=200, corner_radius=0)
        self.sidebar_frame.grid(row=0, column=0, sticky="nsew")
        self.sidebar_frame.grid_rowconfigure(4, weight=1)

        self.logo_label = ctk.CTkLabel(self.sidebar_frame, text="LEGENDARY\nSCALPER", font=ctk.CTkFont(size=20, weight="bold"))
        self.logo_label.grid(row=0, column=0, padx=20, pady=(20, 10))
        
        self.mode_label = ctk.CTkLabel(self.sidebar_frame, text=f"Mode: {'TESTNET' if config.USE_TESTNET else 'REAL'}", 
                                      text_color="orange" if config.USE_TESTNET else "red")
        self.mode_label.grid(row=1, column=0, padx=20, pady=10)

        # Tab Buttons
        self.dashboard_btn = ctk.CTkButton(self.sidebar_frame, text="Dashboard", command=lambda: self.select_frame("dashboard"))
        self.dashboard_btn.grid(row=2, column=0, padx=20, pady=10)
        
        self.console_btn = ctk.CTkButton(self.sidebar_frame, text="Console", command=lambda: self.select_frame("console"))
        self.console_btn.grid(row=3, column=0, padx=20, pady=10)

        # Control Buttons
        self.start_btn = ctk.CTkButton(self.sidebar_frame, text="START BOT", command=self.start_bot, fg_color="green", hover_color="darkgreen")
        self.start_btn.grid(row=5, column=0, padx=20, pady=10)
        
        self.stop_btn = ctk.CTkButton(self.sidebar_frame, text="STOP BOT", command=self.stop_bot, fg_color="red", hover_color="darkred", state="disabled")
        self.stop_btn.grid(row=6, column=0, padx=20, pady=(10, 20))

        # Main Area Frames
        self.dashboard_frame = ctk.CTkFrame(self, corner_radius=0, fg_color="transparent")
        self.console_frame = ctk.CTkFrame(self, corner_radius=0, fg_color="transparent")
        
        self.setup_dashboard()
        self.setup_console()
        
        # Select default
        self.select_frame("dashboard")
        
        # Start UI Update Loop waiting for bot
        self.update_ui()

    def select_frame(self, name):
        # Hide all
        self.dashboard_frame.grid_forget()
        self.console_frame.grid_forget()
        
        # Show selected
        if name == "dashboard":
            self.dashboard_frame.grid(row=0, column=1, sticky="nsew")
        elif name == "console":
            self.console_frame.grid(row=0, column=1, sticky="nsew")

    def setup_console(self):
        self.console_frame.grid_columnconfigure(0, weight=1)
        self.console_frame.grid_rowconfigure(0, weight=1)
        
        self.log_text = ctk.CTkTextbox(self.console_frame, font=("Consolas", 12))
        self.log_text.grid(row=0, column=0, padx=20, pady=20, sticky="nsew")
        self.log_text.configure(state="disabled")
        
        # Setup Logger
        text_handler = TextHandler(self.log_text)
        text_handler.setFormatter(logging.Formatter('%(asctime)s | %(levelname)s | %(message)s', datefmt='%H:%M:%S'))
        
        # Add to root logger to capture everything
        logging.getLogger().addHandler(text_handler)
        logging.getLogger().setLevel(logging.INFO)

    def setup_dashboard(self):
        self.dashboard_frame.grid_columnconfigure((0, 1, 2), weight=1)
        
        # --- Stats Cards ---
        self.card_frame = ctk.CTkFrame(self.dashboard_frame)
        self.card_frame.grid(row=0, column=0, columnspan=3, padx=20, pady=20, sticky="ew")
        self.card_frame.grid_columnconfigure((0, 1, 2), weight=1)

        # Balance
        self.bal_label = ctk.CTkLabel(self.card_frame, text="Balance", font=("Arial", 14))
        self.bal_label.grid(row=0, column=0, pady=(10,0))
        self.bal_value = ctk.CTkLabel(self.card_frame, text="$0.00", font=("Arial", 24, "bold"), text_color="#00ff00")
        self.bal_value.grid(row=1, column=0, pady=(0,10))

        # Active Positions
        self.pos_label = ctk.CTkLabel(self.card_frame, text="Active Positions", font=("Arial", 14))
        self.pos_label.grid(row=0, column=1, pady=(10,0))
        self.pos_value = ctk.CTkLabel(self.card_frame, text="0", font=("Arial", 24, "bold"))
        self.pos_value.grid(row=1, column=1, pady=(0,10))
        
        # Unrealized PnL
        self.pnl_label = ctk.CTkLabel(self.card_frame, text="Unrealized PnL", font=("Arial", 14))
        self.pnl_label.grid(row=0, column=2, pady=(10,0))
        self.pnl_value = ctk.CTkLabel(self.card_frame, text="$0.00", font=("Arial", 24, "bold"))
        self.pnl_value.grid(row=1, column=2, pady=(0,10))

        # --- Positions Table ---
        self.lbl_positions = ctk.CTkLabel(self.dashboard_frame, text="Active Positions", font=("Arial", 18, "bold"), anchor="w")
        self.lbl_positions.grid(row=1, column=0, padx=20, pady=(10, 5), sticky="w")

        self.scroll_frame = ctk.CTkScrollableFrame(self.dashboard_frame, label_text="Symbol | Step | Margin | PnL | Price")
        self.scroll_frame.grid(row=2, column=0, columnspan=3, padx=20, pady=10, sticky="nsew")
        self.dashboard_frame.grid_rowconfigure(2, weight=1)
        
        self.position_rows = {} # Dict to track row widgets

    def start_bot(self):
        if self.bot and self.bot.running:
            return

        self.start_btn.configure(state="disabled")
        self.stop_btn.configure(state="normal")
        self.status_msg("Starting bot...")
        
        self.bot_thread = threading.Thread(target=self.run_bot_thread, daemon=True)
        self.bot_thread.start()

    def run_bot_thread(self):
        try:
            self.bot = LegendaryScalper()
            self.bot.run()
        except Exception as e:
            logging.error(f"Bot Crashed: {e}")
            self.stop_bot()

    def stop_bot(self):
        if self.bot:
            self.bot.stop()
            self.status_msg("Stopping...")
            # Thread will join naturally
        
        self.start_btn.configure(state="normal")
        self.stop_btn.configure(state="disabled")

    def status_msg(self, msg):
        logging.info(f"GUI: {msg}")

    def update_ui(self):
        # Update every 1s
        if self.bot and self.bot.running and self.bot.martingale:
            try:
                status = self.bot.martingale.get_status()
                
                # 1. Update Cards
                positions = status.get('positions', {})
                self.pos_value.configure(text=str(len(positions)))
                
                total_margin = sum(p['total_margin'] for p in positions.values())
                total_pnl = sum(p.get('unrealized_pnl', 0) for p in positions.values())
                
                pnl_color = "#00ff00" if total_pnl >= 0 else "#ff5555"
                self.pnl_value.configure(text=f"${total_pnl:.2f}", text_color=pnl_color)
                
                # We can grab wallet balance from client if available, or just track PnL
                # For now, let's just show Total Margin Used
                self.bal_label.configure(text="Total Margin Used")
                if total_margin > 0:
                   self.bal_value.configure(text=f"${total_margin:.2f}")
                else:
                    self.bal_value.configure(text="$0.00")

                # 2. Update Table
                # Clear old (inefficient but safe for now, better to diff ideally)
                for widget in self.scroll_frame.winfo_children():
                    widget.destroy()

                for sym, pos in positions.items():
                    # Row container
                    row = ctk.CTkFrame(self.scroll_frame)
                    row.pack(fill="x", padx=5, pady=5)
                    row.grid_columnconfigure((0,1,2,3,4), weight=1)
                    
                    # Columns
                    ctk.CTkLabel(row, text=sym, font=("Arial", 12, "bold")).grid(row=0, column=0, pady=5)
                    
                    step_text = f"Step {pos['step']}/{len(config.MARTINGALE_STEPS)}"
                    step_color = "orange" if pos['step'] > 4 else "gray"
                    ctk.CTkLabel(row, text=step_text, text_color=step_color).grid(row=0, column=1)
                    
                    ctk.CTkLabel(row, text=f"${pos['total_margin']:.2f}").grid(row=0, column=2)
                    
                    pnl = pos.get('unrealized_pnl', 0)
                    p_color = "#00ff00" if pnl >=0 else "#ff5555"
                    ctk.CTkLabel(row, text=f"${pnl:.2f}", text_color=p_color).grid(row=0, column=3)
                    
                    ctk.CTkLabel(row, text=f"Avg: {pos['average_entry']:.4f}").grid(row=0, column=4)

            except Exception as e:
                pass # Don't crash UI loop
        
        self.after(1000, self.update_ui)

if __name__ == "__main__":
    app = LegendaryGUI()
    app.mainloop()
