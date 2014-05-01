#!/usr/bin/env python2

# This script migrates data from an SQLite database (clog v0.4)
# to a new PostgreSQL database (clog v0.5). The PostgreSQL database
# should already be created with schema.sql. ALL DATA ON POSTGRESQL
# CLOG DATABASE WILL BE DELETED before data from SQLite be migrated.

SQLITE = "clogd.db"
PG_HOST = "localhost"
PG_DB = "clog"
PG_USER = "clog"
PG_PASSWD = "clog"

print "You should read this script documentation prior to running "
print "it! Press ENTER to migrate data or CTRL-C to abort."
raw_input()

import sqlite3
import psycopg2
import zlib
from hashlib import sha1

print "connecting to SQLite database %s..." % SQLITE
sqlite_conn = sqlite3.connect(SQLITE)

print "connecting to PostgreSQL database %s on %s..." % (PG_DB, PG_HOST)
pg_conn = psycopg2.connect(host=PG_HOST, user=PG_USER,
    password=PG_PASSWD, database=PG_DB)

print "deleting data on PostgreSQL..."
cp = pg_conn.cursor()
cp.execute("truncate table jobconfig, jobconfigalert, jobhistory, jobs, "
    "outputs, sessions, users")
cp.close()

print "migrating jobs output to new format..."
cs = sqlite_conn.cursor()
cs.execute("select output from jobhistory")
for row in cs:
    output = zlib.decompress(row[0]) 
    output_sha1 = sha1(output).hexdigest()
    output = buffer(zlib.compress(output))
    cp = pg_conn.cursor()
    cp.execute("select count(*) from outputs where sha1=%(output_sha1)s",
        locals())
    if cp.fetchone()[0] == 0:
        cp.execute("insert into outputs (sha1, output) "
            "values (%(output_sha1)s, %(output)s)", locals())
    cp.close()
cs.close()

print "migrating users to new format..."
cs = sqlite_conn.cursor()
cs.execute("select * from users")
for row in cs:
    cp = pg_conn.cursor()
    cp.execute("insert into users values (%s, %s, %s)",
        (row[0], row[1], bool(row[2])))
    cp.close()

tables = [("jobconfigalert", 2), ("jobconfig", 2), ("jobs", 8),
    ("jobhistory", 8), ("sessions", 3)]
for table, cols in tables:
    print "migrating table %s..." % table
    cs = sqlite_conn.cursor()
    cs.execute("select * from %s" % table)
    for row in cs:
        cp = pg_conn.cursor()
        row = list(row)
        if table == "jobhistory":
            row[-1] = sha1(zlib.decompress(row[-1])).hexdigest()            
        cp.execute("insert into %s values (%s)" % (table, ",".join(["%s"]*cols)),
            tuple(row))
        cp.close()
    cs.close()

print "updating jobs sequential..."
cp = pg_conn.cursor()
cp.execute("select max(id) from jobs")
row = cp.fetchone()
if row:
    seq = row[0] + 1
    cp.execute("alter sequence jobs_id_seq restart %d" % seq)
cp.close()

print "commiting PostgreSQL data..."
pg_conn.commit()
