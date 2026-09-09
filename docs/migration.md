# Database Migrations

## What a migration is

A migration is a single file with SQL commands that describes one change to the database schema: creating tables, adding a column, creating an index, and so on.

Instead of changing a live database by hand, we store every schema change as a sequence of files. This gives us three things:

- everyone on the team ends up with the same database structure — you just apply the same files;
- there's a history of how the schema changed over time;
- when the project is set up somewhere new (or in Docker), the schema is created automatically, with no manual steps.

## File naming

Migration files live in the `migrations/` folder and are named like this:

```
v001_init_schema.sql
v002_add_something.sql
```

The number at the start of the name sets the **order in which files are applied** — files are applied strictly in increasing order. This matters because some tables reference others (for example, `post` references `user`), and the referenced table has to be created first.

The number is always zero-padded (`0001`, not `v1`) — otherwise, when file names are sorted as strings, `v10` would come before `v2`, breaking the intended order.

**Important rule:** once a migration file has already been applied by someone on the team (i.e. the runner has already run it against a real database), that file must not be edited anymore. Any new schema change is a new file with the next number, not an edit to an old one.

## How this is set up in our project

**`0001_init_schema.sql`** — the schema itself: the `user`, `category`, `post`, `comment`, and `likes` tables, plus indexes on the columns we filter by often (e.g. `post_id` on comments).

**`embed.go`** — using the `//go:embed *.sql` directive, all `.sql` files in the `migrations` folder get baked directly into the compiled binary. Thanks to this, the Docker image doesn't need to copy the migrations folder separately — it's already part of the executable.

**`runner.go`** — the code that applies migrations. The logic is:


**`db.Init()`** — opens the connection to the database file and immediately calls the migration runner. So every time the server starts, the database is automatically brought up to date — nothing needs to be run separately.

## How to add a new migration

1. Create a new file `migrations/000X_description.sql` with the next number in sequence.
2. Write the SQL you need in it (for example, `ALTER TABLE ... ADD COLUMN ...` or `CREATE TABLE ...`).
3. Start the server (or run the runner separately) — the new migration will be applied automatically, while the ones already applied will be skipped.