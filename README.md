<div align="center">
<pre>
███╗   ███╗██╗   ██╗██████╗ ██╗███████╗███╗   ██╗████████╗ ██████╗ ██████╗ 
████╗ ████║╚██╗ ██╔╝██╔══██╗██║██╔════╝████╗  ██║╚══██╔══╝██╔═══██╗██╔══██╗
██╔████╔██║ ╚████╔╝ ██████╔╝██║█████╗  ██╔██╗ ██║   ██║   ██║   ██║██████╔╝
██║╚██╔╝██║  ╚██╔╝  ██╔══██╗██║██╔══╝  ██║╚██╗██║   ██║   ██║   ██║██╔══██╗
██║ ╚═╝ ██║   ██║   ██║  ██║██║███████╗██║ ╚████║   ██║   ╚██████╔╝██║  ██║
╚═╝     ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚══════╝╚═╝  ╚═══╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝
░▒▓█ R O M   S Y N C   F R O M   T H E   F U T U R E   P A S T █▓▒░
</pre>
</div>

<p align="center">
  <img src="https://img.shields.io/badge/WRITTEN_IN-GO-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/ZERO-DEPENDENCIES-ff00ff?style=for-the-badge" alt="Zero Dependencies">
  <img src="https://img.shields.io/badge/CONCURRENT-TRANSFERS-00ffff?style=for-the-badge" alt="Concurrent">
  <img src="https://img.shields.io/badge/RETRO-GAMING-ff6600?style=for-the-badge" alt="Retro Gaming">
</p>

---

<div align="center">
<pre>
╔═════════════════════════════════════════════════════════════════════════════╗
║  >> INITIALIZING NEURAL LINK TO MYRIENT ARCHIVE...                          ║
║  >> SCANNING 200+ ROM VAULTS...                                             ║
║  >> READY TO JACK IN                                                        ║
╚═════════════════════════════════════════════════════════════════════════════╝
</pre>
</div>

## `// WHAT_IS_THIS.exe`

**MYRIENTOR** is a cyberdeck-approved ROM synchronization tool that downloads your favorite retro game collections from the [Myrient Archive](https://myrient.erista.me/). Think of it as `rsync` but designed by someone who grew up in the arcade.

> *"In the neon-lit future of 2024, we still play games from the past."*

<div align="center">
<pre>
┌───────────────────────────────────────────────────────────────────────────┐
│                                                                           │
│   ▀█▀ █▀▀ █▀▀ █ █   █▀ █▀█ █▀▀ █▀▀ █▀                                     │
│    █  ██▄ █▄▄ █▀█   ▄█ █▀▀ ██▄ █▄▄ ▄█                                     │
│                                                                           │
│   ◤ Pure Go - No external dependencies, just raw power                    │
│   ◤ Parallel Downloads - 2 concurrent transfers, maximum throughput       │
│   ◤ Smart Sync - Only downloads what's changed (size + timestamp check)   │
│   ◤ Auto Cleanup - Purges obsolete local files like digital dust          │
│   ◤ Real-time Stats - Watch the bytes flow like rain on a neon sign       │
│   ◤ 200+ Presets - MAME, No-Intro, arcade boards, vintage computers       │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
</pre>
</div>

## `// INSTALL.sh`

### Download Pre-built Binary (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/coltwillcox/myrientor/releases):

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux    | x64          | `myrientor-linux-amd64.tar.gz` |
| Linux    | ARM64        | `myrientor-linux-arm64.tar.gz` |
| macOS    | x64 (Intel)  | `myrientor-darwin-amd64.tar.gz` |
| macOS    | ARM64 (M1+)  | `myrientor-darwin-arm64.tar.gz` |
| Windows  | x64          | `myrientor-windows-amd64.zip` |
| Windows  | ARM64        | `myrientor-windows-arm64.zip` |

```bash
# Linux/macOS example
curl -LO https://github.com/coltwillcox/myrientor/releases/latest/download/myrientor-linux-amd64.tar.gz
tar -xzf myrientor-linux-amd64.tar.gz
./myrientor-linux-amd64 -version
```

### Build from Source

```bash
# Clone the repository from the grid
git clone https://github.com/coltwillcox/myrientor.git
cd myrientor

# Build the executable
go build -o myrientor .

# Check version
./myrientor -version
```

## `// QUICK_START.sh`

```bash
# Edit the config - enable your desired ROM vaults
# Set "sync": true for collections you want
nano remote.json

# Jack in and start the sync
./myrientor
```

## `// CONFIG_MATRIX.json`

Edit `remote.json` to select your targets:

```json
{
  "base_url": "https://myrient.erista.me/",
  "devices": [
    {
      "remote_path": "files/No-Intro/Nintendo - Game Boy/",
      "sync": true,                          // << FLIP THIS SWITCH
      "local_path": "gb"                     // << YOUR LOCAL VAULT
    }
  ]
}
```

> **Note:** The `local_path` folder names in `remote.json` match the [EmulationStation Desktop Edition (ES-DE)](https://es-de.org/) ROM directory structure. See the [ES-DE User Guide](https://gitlab.com/es-de/emulationstation-de/-/blob/master/USERGUIDE.md) for details on supported systems and folder naming conventions.

### Local Settings

Create or edit an optional `local.json` file to customize settings:

```json
{
  "max_concurrent": 4
}
```

| Setting | Description | Default |
|---------|-------------|---------|
| `max_concurrent` | Number of parallel downloads | `2` |

Settings priority: **command-line flags** > **local.json** > **defaults**

### Command-line Flags

| Flag | Description | Example |
|------|-------------|---------|
| `-version` | Show version information | `./myrientor -version` |
| `-concurrent` | Set number of parallel downloads | `./myrientor -concurrent 8` |
| `-sync` | Sync specific device by `local_path` | `./myrientor -sync gb` |

```bash
# Show version
./myrientor -version

# Sync only Game Boy ROMs with 4 parallel downloads
./myrientor -sync gb -concurrent 4
```

### Available Vaults Include:
<div align="center">
<pre>
╭───────────────────────────────────────────────────────────────╮
│  ◈ MAME                    ◈ No-Intro Collections             │
│  ◈ Atari (2600/5200/7800)  ◈ Nintendo (NES/SNES/N64/GB/GBA)   │
│  ◈ Sega (Genesis/Saturn)   ◈ Sony PlayStation                 │
│  ◈ Neo Geo                 ◈ PC Engine / TurboGrafx           │
│  ◈ Commodore 64            ◈ Amiga                            │
│  ◈ MSX                     ◈ And 200+ more...                 │
╰───────────────────────────────────────────────────────────────╯
</pre>
</div>

## `// SYSTEM_OUTPUT.log`

```
Starting sync of 2 device(s) from https://myrient.erista.me/
═══════════════════════════════════════════════════════════════════════

[1/2] Syncing: files/No-Intro/Nintendo - Game Boy/
───────────────────────────────────────────────────────────────────────
✓ Cleaned up 3 obsolete file(s)
↓ Downloading: Pokemon Red (USA).zip
↓ Downloading: Tetris (World).zip
Files: 1437 checked, 42 downloaded, 1395 skipped, 3 deleted
Transfer: 892.45 MiB / 1.20 GiB (74.4%) @ 12.34 MiB/s
Time: 1m23s

✓ Sync complete
═══════════════════════════════════════════════════════════════════════
✓ Sync(s) completed
```

## `// ARCHITECTURE.dat`

<div align="center">
<pre>
                    ┌────────────────────────────────────┐                  
                    │           MYRIENTOR CORE           │                  
                    │         ══════════════════         │                  
                    │                                    │                  
  ┌─────────┐       │  ┌─────────────┐  ┌─────────────┐  │       ┌─────────┐
  │ CONFIG  │──────▶│  │  DIRECTORY  │  │  SYNC       │  │──────▶│  LOCAL  │
  │ JSON    │       │  │  PARSER     │  │  ENGINE     │  │       │  FILES  │
  └─────────┘       │  └──────┬──────┘  └──────┬──────┘  │       └─────────┘
                    │         │                │         │                  
                    │         ▼                ▼         │                  
  ┌─────────┐       │  ┌─────────────────────────────┐   │                  
  │ MYRIENT │◀─────▶│  │    CONCURRENT DOWNLOADER    │   │                  
  │ SERVER  │       │  │   ┌────────┐   ┌────────┐   │   │                  
  └─────────┘       │  │   │ SLOT 1 │   │ SLOT 2 │   │   │                  
                    │  │   └────────┘   └────────┘   │   │                  
                    │  └─────────────────────────────┘   │                  
                    │                                    │                  
                    │  ┌─────────────────────────────┐   │                  
                    │  │        STATS TRACKER        │   │                  
                    │  │  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░ 67% │   │                  
                    │  └─────────────────────────────┘   │                  
                    └────────────────────────────────────┘                  
</pre>
</div>

## `// HOW_IT_WORKS.asm`

```
; PHASE 1: INITIALIZATION
LOAD    local.json           ; Parse the local configuration
LOAD    remote.json          ; Parse the sacred configuration
SCAN    devices[]            ; Count enabled targets
JMP     sync_loop

; PHASE 2: FOR EACH DEVICE
sync_loop:
  FETCH   remote_listing     ; HTML directory scraping (oldschool)
  MKDIR   local_path         ; Ensure local vault exists
  CALL    cleanup_obsolete   ; Purge the digital dead
  SPAWN   goroutines[2]      ; Parallel download threads
  WAIT    completion         ; Semaphore-controlled sync
  PRINT   stats              ; Real-time progress display
  LOOP    next_device

; PHASE 3: VICTORY
PRINT   "✓ Sync(s) completed"
EXIT    0
```

## `// REQUIREMENTS.txt`

<div align="center">
<pre>
╔═══════════════════════════════════════════════════════════════╗
║  • Go 1.22+  (or any version with range-over-int support)     ║
║  • Internet connection to the Myrient grid                    ║
║  • Sufficient disk space for your ROM collection              ║
║  • A love for retro gaming                                    ║
╚═══════════════════════════════════════════════════════════════╝
</pre>
</div>

## `// LICENSE.nfo`

<div align="center">
<pre>
▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄
█                                                           █
█   MIT LICENSE - FREE AS IN FREEDOM                        █
█   Do whatever you want. Just don't blame me.              █
█                                                           █
▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
</pre>
</div>

---

<div align="center">
<pre>
╔═══════════════════════════════════════╗
║                                       ║
║       "INSERT COIN TO CONTINUE"       ║
║                                       ║
║        ┌───┐  ┌───┐                   ║
║        │ 1 │  │ 2 │  PLAYERS          ║
║        └───┘  └───┘                   ║
║                                       ║
╚═══════════════════════════════════════╝
<br>
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░  MYRIENTOR v0.4.0 - SYNC YOUR MEMORIES FROM THE GRID  ░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
</pre>
</div>

---

<p align="center">
  <i>Built with mass amounts of mass nostalgia for the 8-bit era</i>
</p>
