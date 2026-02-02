```

    ███╗   ███╗██╗   ██╗██████╗ ██╗███████╗███╗   ██╗████████╗ ██████╗ ██████╗
    ████╗ ████║╚██╗ ██╔╝██╔══██╗██║██╔════╝████╗  ██║╚══██╔══╝██╔═══██╗██╔══██╗
    ██╔████╔██║ ╚████╔╝ ██████╔╝██║█████╗  ██╔██╗ ██║   ██║   ██║   ██║██████╔╝
    ██║╚██╔╝██║  ╚██╔╝  ██╔══██╗██║██╔══╝  ██║╚██╗██║   ██║   ██║   ██║██╔══██╗
    ██║ ╚═╝ ██║   ██║   ██║  ██║██║███████╗██║ ╚████║   ██║   ╚██████╔╝██║  ██║
    ╚═╝     ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚══════╝╚═╝  ╚═══╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝

          ░▒▓█ R O M   S Y N C   F R O M   T H E   F U T U R E   P A S T █▓▒░
```

<p align="center">
  <img src="https://img.shields.io/badge/WRITTEN_IN-GO-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/ZERO-DEPENDENCIES-ff00ff?style=for-the-badge" alt="Zero Dependencies">
  <img src="https://img.shields.io/badge/CONCURRENT-TRANSFERS-00ffff?style=for-the-badge" alt="Concurrent">
  <img src="https://img.shields.io/badge/RETRO-GAMING-ff6600?style=for-the-badge" alt="Retro Gaming">
</p>

---

```
    ╔══════════════════════════════════════════════════════════════════════╗
    ║  >> INITIALIZING NEURAL LINK TO MYRIENT ARCHIVE...                   ║
    ║  >> SCANNING 200+ ROM VAULTS...                                      ║
    ║  >> READY TO JACK IN                                                 ║
    ╚══════════════════════════════════════════════════════════════════════╝
```

## `// WHAT_IS_THIS.exe`

**MYRIENTOR** is a cyberdeck-approved ROM synchronization tool that downloads your favorite retro game collections from the [Myrient Archive](https://myrient.erista.me/). Think of it as `rsync` but designed by someone who grew up in the arcade.

> *"In the neon-lit future of 2024, we still play games from the past."*

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│   ▀█▀ █▀▀ █▀▀ █ █   █▀ █▀█ █▀▀ █▀▀ █▀                                       │
│    █  ██▄ █▄▄ █▀█   ▄█ █▀▀ ██▄ █▄▄ ▄█                                       │
│                                                                             │
│   ◤ Pure Go - No external dependencies, just raw power                      │
│   ◤ Parallel Downloads - 2 concurrent transfers, maximum throughput         │
│   ◤ Smart Sync - Only downloads what's changed (size + timestamp check)     │
│   ◤ Auto Cleanup - Purges obsolete local files like digital dust            │
│   ◤ Real-time Stats - Watch the bytes flow like rain on a neon sign         │
│   ◤ 200+ Presets - MAME, No-Intro, arcade boards, vintage computers         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## `// QUICK_START.sh`

```bash
# Clone the repository from the grid
git clone https://github.com/coltwillcox/myrientor.git
cd myrientor

# Edit the config - enable your desired ROM vaults
# Set "sync": true for collections you want
nano myrient-devices.json

# Jack in and start the sync
go run main.go

# Or compile your own executable
go build -o myrientor main.go
./myrientor
```

## `// CONFIG_MATRIX.json`

Edit `myrient-devices.json` to select your targets:

```json
{
  "base_url": "https://myrient.erista.me/",
  "devices": [
    {
      "remote_path": "files/No-Intro/Nintendo - Game Boy/",
      "sync": true,                          // << FLIP THIS SWITCH
      "local_path": "/your/roms/gameboy/"    // << YOUR LOCAL VAULT
    }
  ]
}
```

### Available Vaults Include:

```
╭──────────────────────────────────────────────────────────────────╮
│  ◈ MAME                    ◈ No-Intro Collections                │
│  ◈ Atari (2600/5200/7800)  ◈ Nintendo (NES/SNES/N64/GB/GBA)      │
│  ◈ Sega (Genesis/Saturn)   ◈ Sony PlayStation                    │
│  ◈ Neo Geo                 ◈ PC Engine / TurboGrafx              │
│  ◈ Commodore 64            ◈ Amiga                               │
│  ◈ MSX                     ◈ And 200+ more...                    │
╰──────────────────────────────────────────────────────────────────╯
```

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

```
                    ┌────────────────────────────────────┐
                    │           MYRIENTOR CORE           │
                    │         ══════════════════         │
                    │                                    │
  ┌─────────┐       │  ┌─────────────┐  ┌─────────────┐  │       ┌─────────┐
  │ CONFIG  │──────▶│  │  DIRECTORY  │  │  SYNC       │  │──────▶│  LOCAL  │
  │  JSON   │       │  │  PARSER     │  │  ENGINE     │  │       │  FILES  │
  └─────────┘       │  └──────┬──────┘  └──────┬──────┘  │       └─────────┘
                    │         │                │         │
                    │         ▼                ▼         │
  ┌─────────┐       │  ┌─────────────────────────────┐   │
  │ MYRIENT │◀─────▶│  │     CONCURRENT DOWNLOADER   │   │
  │ SERVER  │       │  │   ┌───────┐   ┌───────┐     │   │
  └─────────┘       │  │   │ SLOT 1│   │ SLOT 2│     │   │
                    │  │   └───────┘   └───────┘     │   │
                    │  └─────────────────────────────┘   │
                    │                                    │
                    │  ┌─────────────────────────────┐   │
                    │  │      STATS TRACKER          │   │
                    │  │  ▓▓▓▓▓▓▓▓▓▓░░░░░ 67%        │   │
                    │  └─────────────────────────────┘   │
                    └────────────────────────────────────┘
```

## `// HOW_IT_WORKS.asm`

```
; PHASE 1: INITIALIZATION
LOAD    config.json          ; Parse the sacred configuration
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

```
╔═══════════════════════════════════════════════════════════════╗
║  • Go 1.22+  (or any version with range-over-int support)     ║
║  • Internet connection to the Myrient grid                    ║
║  • Sufficient disk space for your ROM collection              ║
║  • A love for retro gaming                                    ║
╚═══════════════════════════════════════════════════════════════╝
```

## `// LICENSE.nfo`

```
  ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄
  █                                                           █
  █   MIT LICENSE - FREE AS IN FREEDOM                        █
  █   Do whatever you want. Just don't blame me.              █
  █                                                           █
  ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
```

---

```
                    ╔═══════════════════════════════════════╗
                    ║                                       ║
                    ║   "INSERT COIN TO CONTINUE"           ║
                    ║                                       ║
                    ║      ┌───┐  ┌───┐                     ║
                    ║      │ 1 │  │ 2 │  PLAYERS            ║
                    ║      └───┘  └───┘                     ║
                    ║                                       ║
                    ╚═══════════════════════════════════════╝

        ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
        ░  MYRIENTOR v1.0 - SYNC YOUR MEMORIES FROM THE GRID  ░
        ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
```

---

<p align="center">
  <i>Built with mass amounts of mass nostalgia for the 8-bit era</i>
</p>
