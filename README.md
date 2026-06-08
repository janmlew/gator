# gator

`gator` is a command-line RSS feed aggregator. Register an account, follow RSS
feeds, run the aggregator to continuously collect posts into a Postgres
database, and browse the latest posts right in your terminal.

This is the capstone project from the
[Boot.dev](https://www.boot.dev) "Build a Blog Aggregator in Go" course.

## Prerequisites

You'll need two things installed to run `gator`:

- **[Go](https://go.dev/doc/install)** 1.25 or newer — to install and run the CLI.
- **[PostgreSQL](https://www.postgresql.org/download/)** 15 or newer — `gator`
  stores users, feeds, and posts in a Postgres database.

## Installation

Install the `gator` CLI with `go install`:

```bash
go install github.com/janmlew/gator@latest
```

This builds the binary and places it in your `$(go env GOPATH)/bin` directory.
Make sure that directory is on your `PATH` so you can run `gator` from anywhere.

## Database setup

1. Create a Postgres database (named `gator`, for example) and note a
   connection string for it, e.g.:

   ```
   postgres://<user>:<password>@localhost:5432/gator?sslmode=disable
   ```

2. Apply the schema migrations. The migrations live in `sql/schema` and are
   managed with [goose](https://github.com/pressly/goose):

   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   goose -dir sql/schema postgres "<your-connection-string>" up
   ```

## Configuration

`gator` reads its configuration from a JSON file named `.gatorconfig.json` in
your **home** directory (`~/.gatorconfig.json`). Create it with your database
connection string:

```json
{
  "db_url": "postgres://<user>:<password>@localhost:5432/gator?sslmode=disable"
}
```

The `current_user_name` field is managed for you — it's set automatically when
you `register` or `login`, so you don't need to add it by hand.

## Usage

If you installed the binary with `go install`, run commands as `gator <command>`.
While developing inside the repo, you can use `go run . <command>` instead.

```bash
gator register <name>     # create a new user and log in as them
gator login <name>        # switch to an existing user
gator users               # list all users (marks the current one)

gator addfeed <name> <url>  # add a feed and automatically follow it
gator feeds                 # list every feed and who added it
gator follow <url>          # follow an existing feed by URL
gator following             # list the feeds you follow
gator unfollow <url>        # stop following a feed

gator agg <interval>      # run the aggregator loop (e.g. gator agg 1m)
gator browse [limit]      # show recent posts from feeds you follow (default 2)

gator reset               # delete all users (and their feeds/posts) — dev only
```

### Typical workflow

```bash
# 1. Create an account
gator register alice

# 2. Add a feed (you'll automatically follow it)
gator addfeed "Boot.dev Blog" "https://blog.boot.dev/index.xml"

# 3. In a separate terminal, leave the aggregator running. It fetches the
#    oldest-collected feed once per interval and saves new posts to the DB.
#    Be considerate of the servers you're scraping — don't use a tiny interval.
gator agg 1m

# 4. Back in your first terminal, browse the latest posts
gator browse 5
```

`agg` is a long-running process — leave it running in the background and stop it
with `Ctrl+C` when you're done.
