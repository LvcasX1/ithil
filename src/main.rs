//! Ithil - A Terminal User Interface for Telegram
//!
//! This is the main entry point for the Ithil TUI client.

#![allow(clippy::large_futures)]

use std::io;
use std::path::PathBuf;
use std::sync::Arc;

use anyhow::{Context, Result};
use clap::Parser;
use tokio::sync::mpsc;
use tracing::{error, info, Level};
use tracing_appender::rolling::{RollingFileAppender, Rotation};
use tracing_subscriber::{fmt, layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

use ithil::app::{Config, Credentials};
use ithil::cache::new_shared_cache;
use ithil::telegram::TelegramClient;
use ithil::ui::App;

/// Ithil - A Terminal User Interface for Telegram
#[derive(Parser, Debug)]
#[command(name = "ithil")]
#[command(author, version, about, long_about = None)]
struct Cli {
    /// Path to configuration file
    #[arg(short, long, value_name = "FILE")]
    config: Option<PathBuf>,

    /// Enable debug logging
    #[arg(short, long)]
    debug: bool,
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();

    // Load configuration
    let config = Config::load(cli.config.as_deref()).context("Failed to load configuration")?;

    // Validate configuration
    config.validate().context("Invalid configuration")?;

    // Apply theme from config
    ithil::ui::Theme::from_config_str(&config.ui.theme).apply();

    // Set up logging
    setup_logging(&config, cli.debug)?;

    info!("Starting Ithil v{}", env!("CARGO_PKG_VERSION"));
    info!("Configuration loaded successfully");

    // Ensure required directories exist
    config
        .ensure_directories()
        .context("Failed to create application directories")?;

    // Run the TUI application
    run_app(config).await
}

/// Set up tracing/logging infrastructure
fn setup_logging(config: &Config, debug: bool) -> Result<()> {
    let log_level = if debug {
        Level::DEBUG
    } else {
        match config.logging.level.to_lowercase().as_str() {
            "trace" => Level::TRACE,
            "debug" => Level::DEBUG,
            "warn" | "warning" => Level::WARN,
            "error" => Level::ERROR,
            // Default to INFO for "info" and any other value
            _ => Level::INFO,
        }
    };

    // Create log directory if it doesn't exist
    if let Some(parent) = config.logging.file.parent() {
        std::fs::create_dir_all(parent).context("Failed to create log directory")?;
    }

    // Set up file appender for logging
    let file_appender = RollingFileAppender::new(
        Rotation::DAILY,
        config.logging.file.parent().unwrap_or(&PathBuf::from(".")),
        config.logging.file.file_name().map_or_else(
            || "ithil.log".to_string(),
            |s| s.to_string_lossy().to_string(),
        ),
    );

    let (non_blocking, guard) = tracing_appender::non_blocking(file_appender);

    // Build the subscriber with file logging
    let env_filter =
        EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new(log_level.as_str()));

    tracing_subscriber::registry()
        .with(env_filter)
        .with(fmt::layer().with_writer(non_blocking).with_ansi(false))
        .init();

    // We need to keep the guard alive, but since we're running until exit,
    // we'll leak it intentionally to keep logging working
    std::mem::forget(guard);

    Ok(())
}

/// Run the main TUI application
async fn run_app(config: Config) -> Result<()> {
    // Set up terminal
    crossterm::terminal::enable_raw_mode().context("Failed to enable raw mode")?;

    let mut stdout = io::stdout();
    crossterm::execute!(
        stdout,
        crossterm::terminal::EnterAlternateScreen,
        crossterm::event::EnableMouseCapture
    )
    .context("Failed to set up terminal")?;

    let backend = ratatui::backend::CrosstermBackend::new(stdout);
    let mut terminal = ratatui::Terminal::new(backend).context("Failed to create terminal")?;

    // Create shared cache
    let cache = new_shared_cache(config.cache.max_messages_per_chat);

    // Get API credentials
    let credentials = Credentials::from_config(&config);

    info!(
        "Using session file: {}",
        config.telegram.session_file.display()
    );

    // Create Telegram client
    let telegram = Arc::new(TelegramClient::new(
        credentials.api_id,
        credentials.api_hash,
        config.telegram.session_file.to_string_lossy().to_string(),
        cache.clone(),
    ));

    // Create update channel for streaming Telegram updates to the UI
    let (update_tx, update_rx) = mpsc::channel(100);
    telegram.set_update_channel(update_tx).await;

    // Create the app
    let mut app = App::new(config, telegram.clone(), cache);
    app.set_update_receiver(update_rx);

    // Spawn Telegram connection in background so UI can render
    let telegram_for_connect = telegram.clone();
    let connect_handle = tokio::spawn(async move {
        info!("Connecting to Telegram...");
        let result = telegram_for_connect.connect().await;
        match &result {
            Ok(()) => info!("Connected to Telegram successfully"),
            Err(e) => error!("Failed to connect to Telegram: {e}"),
        }
        result
    });

    // Run the async event loop with connection happening in background
    let result = app
        .run_async_with_connection(&mut terminal, connect_handle)
        .await;

    // Disconnect from Telegram gracefully
    if telegram.is_connected().await {
        if let Err(e) = telegram.disconnect().await {
            error!("Error disconnecting from Telegram: {e}");
        }
    }

    // Restore terminal
    crossterm::terminal::disable_raw_mode().context("Failed to disable raw mode")?;

    crossterm::execute!(
        terminal.backend_mut(),
        crossterm::terminal::LeaveAlternateScreen,
        crossterm::event::DisableMouseCapture
    )
    .context("Failed to restore terminal")?;

    terminal.show_cursor().context("Failed to show cursor")?;

    result
}
