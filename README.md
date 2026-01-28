# DMarket-To-Telegram Transactions Poster

A Go application to post DMarket Sales, Purchases and Closed Targets transactions to Telegram channel.

## Structure

### Project structure

- `cmd/transactionTracker/`: Main executable entrypoint.
- `services/`: Shared Go modules (API headers, transaction logic, timestamp management).
- `types/`: Data structures and API response definitions.
- `config/`: Configuration files and templates (real keys are ignored by .gitignore).

### Message structure

`Action` `status`

`item name`


`Float: float value` (hidden if item have no float value)

`Phase: string value` (hidden if not applicable)

`Pattern: int value` (hidden if item have no pattern)


`Change: float value $` (amount spent or gained in this transaction)

`Profit: float value $ ` (calculated net profit, only if action = Sell and purchase was found)

`Balance: float value $` (approximate **usable** user balance, see Note 2.2)

## Setup

Requires **Go 1.16+** (tested on Go 1.24.5, Windows 10).

1. Clone the repo: `git clone https://github.com/cyberbebebe/dmarket-transactions-poster.git`
2. `cd dmarket-transactions-poster`
3. Copy and fill config templates:

   PRIVATE KEYs:
   `copy config\secretKeys.example.json config\secretKeys.json` (Windows)

   or `cp config/secretKeys.example.json config/secretKeys.json` (Unix/Mac)

   fill with your DMarket PRIVATE API key(s) (use array even if 1 account) and Telegram bot token(s) (from [@BotFather](https://t.me/BotFather) in Telegram).

   PUBLIC KEYs -> Telegram chat IDs:

   `copy config\chatids.example.json config\chatids.json` (Windows)

   or `cp config/chatids.example.json config/chatids.json` (Unix/Mac)

   fill with your DMarket PUBLIC API key(s) and corresponding Telegram chat ID(s).
   To get a chat ID: Open [web.telegram.org](https://web.telegram.org), go to your channel, and check the URL. If it ends in `#-721752185`, your chatID is `-100721752185`:
   `{"public key here": {"transactions": "-100721752185"}}` in `chatids.json`.

4. Install dependencies: `go mod tidy`
5. Run the app:
   - `go run cmd/transactionTracker/main.go`

     or (i recommend) build an .exe file:

   - `go build -o DMTransactions.exe ./cmd/transactionTracker`

**Troubleshooting**: If configs fail to load, check JSON format. Set `GOPATH` if not default. For multi-account, ensure arrays in JSON match (e.g., private keys index to public keys).

## Dependencies

- Go 1.16+
- [Telegram Bot API](https://github.com/go-telegram-bot-api/telegram-bot-api) v5.5.1

## Examples

Sold with trade-protected status post example:

![Sell trade_protected example](images/sell-tp.png)

(important balance note below)

Target Closed with trade-protected status post example:

![Target Closed trade_protected example](images/target-closed-tp.png)

Reverted Sell post example (from older code version):

![Reverted sell example](images/reverted-sell.png)

(For multi-account setup, the programm cycles through keys)

## Important notes before use:

1. This is **half-vibecoded project** by Go beginner amateur **for self usage**.

   This code is **not** what professional project should be like.

2. There are some notes and inconveniences due to DMarket's web history and dumb "programmer":

   2.1) This code uses web `/history` endpoint with sorting by **updatedAt**. This means that transactions that were trade protected **will be posted again** with the new status "Success" or "Reverted". This "double-posting" can be "fixed" by adding `&sortBy=createdAt` http param to the endpoint in `func GetLastTransactions()`.

   **However,** using `&sortBy=createdAt` will **not** let you know if a transaction got "Reverted" or moved from pending to "Success".

   2.2) Balance amount **is approximation** if transaction type is "Sell" or "Target Closed" due to DMarket's balance field calculated as `usable balance + lot sold price` (**not net income**), so I simulate the 2% fee.
   **However:**
   - 2.2.1) For items cheaper than ~ 7$, the sale fee is 10%.
   - 2.2.2) I do **not** track "Instant sale" and "Trade" actions because i don't use them.

   _Adding new actions tracking and requesting balance requires new functions and changing other._

   2.3) I use `/user-targets/closed` endpoint to check buy history before the main cycle starts.

   There is a commented line: `// totalLimit := 100000`. This variable can be used to limit how many previous **buy** (_purchased_ and _target closed_) **transactions** you want to request in total. For using buy tracking limit:
   - Uncomment `totalLimit` variable line by deleting `// `
   - Change this line `if response.Cursor == "" {` to this:

   `if response.Cursor == "" || len(allTransactions[key]) > totalLimit {`
   - Note that limiting the buy history for one account among a group of accounts requires code changes that are not represented here.

   2.4) Default settings requests up to 10 last updated transactions, with a frequency of 15 seconds. **However**:

   - 2.4.1) This can be modified by increasing the limit from `&limit=10` to `&limit=100` (or any other) in func GetLastTransactions() endpoint or/and changing the timing in main.go: `time.Sleep(15 * time.Second)` for something like `time.Sleep(5 * time.Minute)`.
   - 2.4.2) At trade unlock time (8:00 GMT), DMarket verifies the status of trades and pushes a bunch of transactions to the top of the history. This means there may be many posts at that time. **Critical:** If more transactions happen during your `time.Sleep()` period than your `limit` allows (e.g., 15 transactions happen but limit is 10), the **older** transactions will be **ignored**. To handle this properly, read and follow the instructions in notes **2.4.1** and **2.1**.

   _This "ignoring" behavior could be fixed, but it requires significant changes to the transaction.go service._

   **Alert:** Telegram **can** mute your bot/channel up to 1 minute if you spam too many messages in a few seconds (e.g., 15 messages per 1 second).

   2.5) This code uses **1 telegram bot** for posting across all channels. It can be changed, but this requires some **major changes** in `secretKeys.json` and some `services/` files, so **i do not recommend changing it** unless you have patience to read and modify this not-very-well-written code.

   2.6) This code does **not** print stickers info (applied on skins). Maybe i will add this later.

3. You can suggest any feature/bug/question/idea. For example i'm thinking of:

   3.1) Calculating profit and prices for non-USD currencies. (Easy)

   3.2) Printing the Pending balance next to the "Usable" balance in the message. (Medium)

   3.3) Fetching CSFloat's buy history (using API key) to calculate profit for items bought on CSFloat. (Hard)

   3.4) Reworking the whole project for **combined cross-platform** (CSFloat & DMarket) transaction posting. (Needs **major** work)
   

### TL;DR (Quick Summary)

1. Status: Experimental/Amateur vibecoded project. Expect quirks.

2. Double Posts: Transactions appear twice (Pending → Success) because of DMarket's sorting logic. Check note **2.1** for details.

3. Money: Balance is an estimate. It hardcodes a 2% fee and misses the 10% fee for cheap items.

4. Spam Warning: Watch out for the 8:00 GMT unlock wave; high volume might trigger a Telegram mute. Adjust timing and request limits. Check note **2.4.2** for details.

5. Architecture: Uses one Telegram bot to manage all channels.
