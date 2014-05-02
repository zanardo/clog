`clog` is a centralized monitoring tool for crontabs, scheduled events, etc,
running on other computers. The client executes a script on a Linux or Windows
machine, and sends the result to the server, which shows jobs in a web based
interface, with start and end time, duration, status (success or failure),
script output, etc. The server can send notification alerts by e-mail when jobs
fail.

# Server installation

clog server is developed in Python, needs PostgreSQL and is tested only in
Linux.

    git clone https://github.com/zanardo/clog.git
    cd clog
    make

Create a new user in PostgreSQL, which will own the database. The user should
have login role, and should not have superuser, create database or other
privileges. Sample:

    createuser -ClRS clog

Create a new database in PostgreSQL, owned by the new user:

    createdb -E UTF-8 -l en_US.UTF-8 -O clog -T template0 clog

Import the initial database schema:

    psql -U clog clog < schema.sql

Configure clogd:

    cp config_example.yaml config.yaml
    $EDITOR config.yaml

You can run `clogd` with `make run-server` or start if with supervisor:

    ./.venv/bin/python clogd config.yaml

After the installation, access with a browser, with user `admin` and password
`admin`, create a new user for you, and delete the `admin` user.

# Client installation

clog client is developed in Go and is tested on Linux and Windows.

Just compile `clog/clog.go` with Go in Linux or Windows:

    go build clog.go

You can copy `clog` executable (`clog.exe` on Windows) somewhere on your
`$PATH` and it is done.

# Client usage

Run `clog` one time so it can create its directories:

    clog

Now you can create a script inside `$HOME/.clog-scripts`. For example, the
script name could be `test.sh`.

To run the script:

    clog run test.sh

clog will put the result into a queue inside `$HOME/clog-queue`, and this queue
should be dispatched to the server:

    clog send-queue http://servername:7890/

You can schedule both actions on cron or Windows task scheduler.
