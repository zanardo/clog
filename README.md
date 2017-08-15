`clog` is a centralized monitoring tool for crontabs, scheduled events, etc,
running on other computers. The client executes a script on a Linux machine,
and sends the result to the server, which shows jobs in a web based interface,
with start and finish time, duration, status (success or failure), script
output, etc. The server can send notification alerts by e-mail when jobs fail.

# Server installation

`clog` server is developed in Python, needs PostgreSQL and is tested only in
Linux.

```bash
git clone https://github.com/zanardo/clog
cd clog
make dev
```

Create a new user in PostgreSQL, which will own the database. The user should
have login role, and should not have superuser, create database or other
privileges. Sample:

```bash
createuser -P clog
```

Create a new database in PostgreSQL, owned by the new user:

```bash
createdb -E UTF-8 -l en_US.UTF-8 -O clog -T template0 clog
```

Import the initial database schema:

```bash
psql -U clog clog < schema.sql
```

Configure `clogd`:

```bash
cp config.yml.example config.yml
$EDITOR config.yml
```

You can run `clogd` with `make run` or start if manually:

```
export CLOGD_CONF=$(pwd)/config.yml
.venv/bin/waitress-serve --host 0.0.0.0 --port 7890 clogd:app
```

After the installation, access with a browser, with user `admin` and password
`admin`, create a new user for you, and delete the `admin` user.

# Client installation

`clog` client is written in Python 2.7 and does not need any special module more
than the Python standard library. You can install from source distribution:

```bash
make install-cli
```

# Client usage

Run `clog` one time so it can create its directories:

```bash
clog
```

Now you can create a script inside `$HOME/.clog-scripts`. For example, the
script name could be `test.sh`.

To run the script:

```bash
clog run test.sh
```

`clog` will put the result into a queue inside `$HOME/.clog-queue`, and this
queue should be dispatched to the server:

```
clog send-queue http://servername:27890/
```

You can schedule both actions on cron.
