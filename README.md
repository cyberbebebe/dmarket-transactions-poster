# DMarket-To-Telegram Transactions Poster

A Go application to post DMarket Sales, Purchases and Closed Targets transactions to Telegram channel.

## Structure

### Project structure

- `cmd/transactionTracker/`: Main executable entrypoint.
- `services/`: Shared Go modules (API headers, transaction logic, timestamp management).
- `types/`: Data structures and API response definitions.
- `config/`: Configuration file template (real keys are ignored by .gitignore).

### Message structure

`[Action]` `[Status]`

`Item Name`

`Float: 0.1842...` (Hidden if item has no float)

`Phase: Phase 2` (Hidden if item has no phase)

`Pattern: 123` (Hidden if item has no pattern)

`Change: + 600.00 $` (Amount spent or gained)

`Profit: + 100.00 $` (Hidden if buy price not found. Shows "/ +20.00 %" if profit_percent = true)

`Balance: 500.00 $` (Usable balance. Shows "/ pending $" if advanced_balance = true)

## Setup

### Quick Start (Pre-built)

1. Download `DmarketTrackerV*.zip` from the [Releases page](https://github.com/cyberbebebe/dmarket-transactions-poster/releases)
2. Read the `Use guide.txt` inside the archive.
3. Run `DmarketTracker.exe`

### Advanced (Build from source)

Requires **Go 1.21+** (tested on Go 1.24.5, Windows 10).

1. Clone the repo:

   `git clone https://github.com/cyberbebebe/dmarket-transactions-poster.git`

   `cd dmarket-transactions-poster`

2. Copy and fill config:

   Windows: `copy config\config.example.json config\config.json`

   Unix/Mac: `cp config/config.example.json config/config.json`

   Open config/config.json and fill in your accounts and telegram data. The file must start with `[` and end with `]`. Example file is inside config folder.

   Fields Guide:
   - dmarket_key: (Required) Your Private API key from DMarket.
   - csfloat_key: Your CSFloat API (dev) Key. (optional, leave empty "" if not used)
   - telegram_token: Get this from @BotFather.
   - telegram_chat_id: Your channel or group ID.
     Open web.telegram.org, go to your channel, and check the URL. If it ends in `#-721752185`, your chatID is `-100721752185`.
   - advanced_balance: Set to true (recommended) to show pending balance (e.g., / 271.2 $).
   - profit_percent: Set to true (recommended) to show profit percentage (e.g., / + 7.52%).
   - ignore_released: Set to true (recommended) to ignore transactions that changed status from "trade_protected" to "success" ("Reverted" transactions will still be posted)

3. Install dependencies: `go mod tidy`

4. Run the app:
   Directly
   - `go run cmd/transactionTracker/main.go`

   Build .exe (Recommended):
   - `go build -o DmarketTracker.exe ./cmd/transactionTracker`

**Troubleshooting**:

- **App crashes immediately?**. Run it via the terminal (cmd or PowerShell) to see the error message.
- **JSON Error?** Ensure your `config.json` has commas `,` between fields and account blocks, but no comma after the last field/block.
- **CSFloat not syncing?** The auto-updater runs on app start and then every **3 days**. Check if your API key is valid.

## Dependencies

- Go 1.21+
- [Telegram Bot API](https://github.com/go-telegram-bot-api/telegram-bot-api) v5.5.1

## Examples

Sold with trade-protected status post example:

![Sell trade_protected example](images/sell-tp.png)

Target Closed with trade-protected status post example (from older code version):

![Target Closed trade_protected example](images/target-closed-tp.png)

Reverted Sell post example (from older code version):

![Reverted sell example](images/reverted-sell.png)

(For multi-account setup, the programm cycles through keys)

## Important notes before use:

1. This is **almost fully** vibecoded project by Go beginner amateur **for self usage**.

   This code is **not** what professional project should be like.

2. There are some notes due to DMarket's web history, Telegram rate limits and dumb programmer:

   2.1) This code uses web `/history` endpoint with sorting by **updatedAt** (default). This means that transactions that were trade protected **will be posted again** with the new status "Success" or "Reverted". This "double-posting" can be "fixed" by
   - adding `&sortBy=createdAt` http param to the endpoint in `func FetchNewTransactions()`

   or
   - setting "ignore_released" to "true" in account config (recommended).

   **However,** using 1st fix method (`&sortBy=createdAt`) will **not** let you know if a transaction got "Reverted" or moved from pending to "Success".

   2.2) Default settings requests up to 50 _last updated_ transactions, with a frequency of 15 seconds. **However**:
   - This can be modified by decreasing the limit from `&limit=50` to `&limit=10` (or any other) in func FetchNewTransactions() endpoint or/and changing the timing in main.go: `time.Sleep(15 * time.Second)` for something like `time.Sleep(5 * time.Minute)`.
   - At trade unlock time (8:00 GMT), DMarket verifies the status of trades and pushes a bunch of transactions to the top of the history. This means there may be many posts at that time. **Critical:** If more transactions happen during your `time.Sleep()` period than your `limit` allows (e.g., 15 transactions happen but limit is 10), the **older** transactions will be **"ignored"**. To handle this, use higher `&limit=` and follow the instructions in notes **2.1** _(about "ignore_released")_

   _This "ignoring" behavior could be fixed by using queue for messages, but i recommend setting "ignore_released" to "true"._

   **Alert:** Telegram **can** mute your bot or/and channel up to 1 minute if you spam too many messages in a few seconds (e.g., 25 messages per 2 second).

   2.3) This code does **not** print stickers info (applied on skins). Maybe i will add this later.

3. You can ask anything or suggest any feature/bug/idea.
