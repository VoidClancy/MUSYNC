# MUSYNC

MUSYNC is an automated synchronization tool designed to mirror online music playlists (such as Spotify public shares) onto Deezer. Unlike simple title-based search scripts, MUSYNC queries track records globally by **ISRC (International Standard Recording Code)** to ensure exact matched recordings are synchronized.

It is structured to run as a fast, state-preserving executable—perfect for scheduled executions (cron jobs) on cloud platforms or continuous integration pipelines like GitHub Actions.

## Key Features

- **ISRC Resolution**: Synchronizes tracks using industry-standard ISRC codes to guarantee exact recording matches.
- **Concurrent Execution with Intelligent Throttling**: Utilizes Go's concurrency primitives with an active rate limiter capped at 9 requests per second to avoid triggering Deezer's quota limits.
- **State Preservation**: Saves resolved playlist IDs directly to a local configurations file (`playlists.json`).
- **Persistent Local Cache**: Maintains a local `isrc_cache.json` tracking resolved mappings of ISRC codes to Deezer IDs and track titles. Subsequent runs read from the cache first, resulting in zero API requests for already-resolved tracks.
- **Telegram Notification Integration**: Dispatches rich, colorized execution summaries (indicating elapsed duration, tracks processed/added, and error status) straight to a Telegram bot.
- **GitHub Actions Ready**: Fully configured workflow to run automatically every 4 hours, commit updated playlist IDs and the ISRC cache back to the repository, and send status updates.

## System Architecture

The following diagram illustrates how MUSYNC synchronizes track playlists from source web pages to Deezer.

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
