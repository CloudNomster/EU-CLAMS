# EU-CLAMS

This is a Go-based tool for tracking and analyzing Entropia Universe global events. It monitors your chat log in real-time, capturing information about globals, Hall of Fame entries, and other significant events.

## Features

- Real-time monitoring of chat logs for global events
- Support for both player and team globals
- Tracking of various global types:
  - Kill globals (player and team)
  - Crafting globals
  - Mining/deposit finds
- Hall of Fame (HoF) detection
- Automatic screenshots of globals and HoFs
- Detailed statistics and analysis
- Web server for viewing statistics in a browser

## Project Structure

- `cmd/app`: Application entry points
- `pkg`: Reusable utility packages
- `internal`: Core application packages
  - `config`: Configuration management
  - `logger`: Logging functionality
  - `stats`: Statistics generation
  - `storage`: Data persistence and chat log processing
- `src`: Service and business logic
  - `service`: Core services for data processing and statistics

## Getting Started

### Installation

1. Clone the repository
2. Install prerequisites:
   - For GUI mode: Install MinGW-w64 or TDM-GCC for Windows (download from https://jmeubank.github.io/tdm-gcc/)
3. Build the application:
   - Using the build script: `build.bat`
   - Manual build: `go build -o eu-clams.exe ./cmd/app`

### Configuration

Configuration can be provided in three ways:

1. Command-line flags:
   ```bash
   eu-clams -player "YourCharacterName" -team "YourTeamName"
   ```

2. Configuration file (automatically loaded if present):
   - Copy `config.example.yaml` to `config.yaml` and edit as needed
   - The application will automatically load `config.yaml` from the current directory
   ```yaml
   # config.yaml
   app_name: EU-CLAMS
   database_path: ./data/db.yaml
   player_name: YourCharacterName
   team_name: YourTeamName
   enable_screenshots: true
   screenshot_directory: ./data/screenshots
   game_window_title: Entropia Universe Client
   enable_web_server: false
   web_server_port: 8080
   ```

3. GUI Configuration Dialog (when using GUI mode):
   - Launch the application: `eu-clams`
   - Click the "Configure" button to open the configuration dialog
   - Enter your settings and click "Save Configuration"
   - Settings will be saved to `config.yaml` for future use

### Command-line Options

```
-config string           Path to configuration file
-log string              Path to chat log file (default: Documents\Entropia Universe\chat.log)
-player string           Your character name
-team string             Your team name
-import                  One-time import without monitoring
-stats                   Show statistics for your globals
-monitor                 Monitor chat log for changes
-version                 Display version information
-cli                     Use command-line interface instead of GUI
-screenshots bool        Enable screenshots for globals and HoFs (default: true)
-screenshot-dir string   Directory to save screenshots (default: ./data/screenshots)
-game-window string      Game window title (default: Entropia Universe Client)
-web bool                Start a web server to view statistics (default: false)
-web-port int            Port for the web server (default: 8080)
```

### Usage Modes

#### 1. Graphical User Interface (GUI) Mode (Default)
```bash
eu-clams
```
- Provides a user-friendly interface for all operations
- Configure player and team names
- Start/stop monitoring with a button click
- Import chat logs via file browser
- View statistics in a formatted window
- Real-time status updates

#### 2. Command-line Interface (CLI) Mode
```bash
eu-clams -cli
```

This mode offers several sub-modes:

##### a. Real-time Monitoring
```bash
eu-clams -cli -player "YourCharacterName" -team "YourTeamName"
```
- Automatically watches chat log for new globals
- Processes any new entries in real-time
- Updates database immediately when new globals are found
- Shows live feedback for new entries
- Press Ctrl+C to stop monitoring

##### b. One-time Import
```bash
eu-clams -cli -import -player "YourCharacterName" -team "YourTeamName"
```
- Processes the entire chat log once
- Shows progress during import
- Exits after completion
- Useful for initial setup or catching up after being offline

##### c. Statistics View
```bash
eu-clams -cli -stats -player "YourCharacterName"
```
Displays comprehensive statistics about your globals:
- Total globals and HoFs
- Total PED value
- Highest value global
- Breakdown by type (kills/crafting/mining)
- Location statistics
- Team contribution analysis
- Time-based analysis
- Last update timestamp

##### d. Web Server View
```bash
eu-clams -cli -web -player "YourCharacterName" -web-port 8080
```
Starts a web server that displays your statistics and globals:
- Interactive, mobile-friendly web interface
- Live updates via WebSockets when new globals are detected
- Separate views for globals and Hall of Fame entries
- JSON API endpoints for integration with other tools
- Dark/light mode toggle
- Accessible from any device on your network

### Data Storage

The tool uses a YAML database file to store all global information:

```yaml
# Default location: ./data/db.yaml
globals:
  - timestamp: 2025-05-16T10:00:00Z
    type: kill
    player: YourCharacterName
    target: CreatureName
    value: 100
    location: LocationName
    is_hof: true
    raw_message: Original chat log message
```

Key features:
- Automatic database creation and updates
- Maintains original chat log messages
- Tracks last processed position to avoid duplicates
- Supports both relative and absolute paths

### Screenshots

The tool can automatically take screenshots when you or your team gets a global or Hall of Fame entry:

- Screenshots are saved in the configured directory (default: `./data/screenshots`)
- Filename format: `[global/hof]_[type]_[player/team]_[timestamp].png`
- Only triggered for relevant globals (your player or team name)
- All Hall of Fame entries trigger screenshots
- Can be enabled/disabled in configuration

To configure screenshots:
1. Go to "Configure" in the app
2. Enable/disable "Enable Screenshots" option
3. Set "Screenshot Directory" to your preferred location
4. Set "Game Window Title" if your game window has a different title. The default should work. (default: "Entropia Universe Client")

#### How Screenshots Work

When a global or Hall of Fame event is detected:
1. The app checks if screenshots are enabled
2. It looks for a window with the specified title ("Entropia Universe Client" by default)
3. It takes a screenshot of that window
4. The screenshot is saved to the configured directory with a filename that includes:
   - Event type (global or HoF)
   - Global type (kill, craft, etc.)
   - Player or team name
   - Timestamp

### Web Server

EU-CLAMS includes a built-in web server that provides a browser-based dashboard for viewing your global and Hall of Fame statistics:

#### Features:
- Real-time updates via WebSockets when new globals are detected
- Mobile-friendly responsive design
- Dark/light mode toggle
- Sortable and filterable tables of globals and HoF entries
- Summary statistics and charts
- Direct links to screenshots (if enabled)

#### Configuration:

The web server can be configured in three ways:

1. Configuration file:
   ```yaml
   # In config.yaml
   enable_web_server: true
   web_server_port: 8080
   ```

2. Configuration GUI:
   - Open the configuration dialog
   - Check "Enable Web Server"
   - Set the desired port number
   - Save the configuration

3. Command-line flags:
   ```bash
   eu-clams -web -web-port 9090
   ```

#### Starting the Web Server:

1. **From GUI:**
   - Click the "Launch Web Dashboard" button in the main interface
   - A browser window will automatically open with the dashboard

2. **From command line:**
   ```bash
   eu-clams -cli -web -player "YourCharacterName"
   ```

3. **Automatically on startup:**
   - Set `enable_web_server: true` in your config.yaml
   - Start the application normally

#### API Endpoints:

The web server provides several API endpoints:

- `/api/stats` - Get summary statistics
- `/api/globals` - Get all globals
- `/api/hofs` - Get all Hall of Fame entries
- `/ws` - WebSocket endpoint for real-time updates

Example filename: `hof_kill_YourName_2025-05-16_10-00-00.png`

#### Command-line Screenshot Control

You can also control screenshot behavior via command line:
```bash
# Disable screenshots
eu-clams -screenshots=false

# Change screenshot directory
eu-clams -screenshot-dir="C:\MyScreenshots"

# Change game window title
eu-clams -game-window="My Entropia Client"
```

### Examples

1. Launch with GUI (default):
   ```bash
   eu-clams
   ```

2. Use command-line mode to monitor chat logs:
   ```bash
   eu-clams -cli -player "YourName" -team "YourTeam"
   ```

3. Import existing chat log and view statistics in CLI mode:
   ```bash
   eu-clams -cli -import -player "YourName" -stats
   ```

4. Process a specific chat log file in CLI mode:
   ```bash
   eu-clams -cli -log "C:\Games\EntrU\chat.log" -player "YourName"
   ```

5. Use a custom configuration location:
   ```bash
   eu-clams -config "path/to/config.yaml"
   ```

### Tips

1. The tool automatically finds your chat log in the default location
2. Use the -import flag for initial setup, then run in monitor mode
3. Regular statistics checks help track your progress
4. Team name is optional but recommended for complete tracking
5. The tool safely handles server restarts and game crashes
