# MUSYNC

MUSYNC is a CLI tool written in Go that synchronizes music playlists to Deezer using ISRC (International Standard Recording Code) identifiers.

It runs as a single-pass command designed to be scheduled as a cron job or run in a GitHub Actions workflow.

## Features

- **ISRC Matching**: Resolves tracks by ISRC.
- **Rate Limiting**: Throttles requests to 9 requests per second.
- **Config Updates**: Updates playlist IDs in `playlists.json`.
- **ISRC Cache**: Caches resolved mappings in `isrc_cache.json`.
- **Telegram Notifications**: Sends run summaries.
- **GitHub Actions**: Runs sync on a cron schedule.

## System Architecture

The following diagram illustrates the execution flow:

```mermaid
graph TD
    subgraph Input Configuration
        A["playlists.json"] -->|1. Load Playlists| B["MUSYNC Engine"]
    end

    subgraph ISRC Parsing
        B -->|2. Scraping Trigger| C["isrcHunt Scraper"]
        C -->|3. HTTP Request| D["Web Source"]
        D -->|4. HTML Content| C
        C -->|5. Extract & Return ISRCs| B
    end

    subgraph Resolution Pipeline
        B -->|6. Lookup ISRC| E["ISRC Cache Loader"]
        E -->|7. Check Cache| F{"isrc_cache.json"}
        F -->|8a. Cached Match Found| B
        F -->|8b. Not Cached| G["Rate-Limited Worker Pool"]
        G -->|9. Query ISRC| H["Deezer API"]
        H -->|10. Return ID & Title| G
        G -->|11. Update Cache File| F
        G -->|12. Return Resolved ID| B
    end

    subgraph Deezer Integration
        B -->|13. Initialize| I["Deezer Session Manager"]
        I -->|14. Get/Create Playlist| J["Deezer Playlists"]
        I -->|15. Check Tracks| J
        I -->|16. Add New Songs| J
    end

    subgraph Notifications & State Commit
        B -->|17. Save Updated IDs| K["playlists.json Updater"]
        B -->|18. Trigger Telegram Alert| L["Telegram Client"]
        L -->|19. Send Notification| M["Telegram Chat"]
    end

    style B fill:#3b82f6,stroke:#1d4ed8,stroke-width:2px,color:#fff
    style G fill:#10b981,stroke:#047857,stroke-width:2px,color:#fff
    style J fill:#f59e0b,stroke:#d97706,stroke-width:2px,color:#fff
```
