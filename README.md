This script processes cs demo files to extract ninja defuses

## Configuration

Set these variables in `main.go`:

- **`demoDirectory`**: Path to your `.dem` files.
  ```go
  demoDirectory := "D:\\SteamLibrary\\steamapps\\common\\Counter-Strike Global Offensive\\game\\csgo\\replays"
  ```
- **`playerID`**: SteamID64 of the target player.
  ```go
  playerID := uint64(76561199121731119)
  ```
- **`filterDate`**: Process files modified on or after this date (format: `YYYY-MM-DD`).
  ```go
  filterDate := "2024-01-01"
  ```
- **`maxParallelism`** (optional): Number of concurrent file processes. Default: 25.
- **`defuseOutputFile`** (optional): Output file name. Default: `defuse_results.txt`.

Example:
```
File: demo1.dem, Date: 2024-01-15, Map: de_dust2, Final Score: CT 16 - T 13, Round: 12, Enemies Alive: 3
```

