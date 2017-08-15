# -*- coding: utf-8 -*-
#
# Copyright (c) 2013-2017, Antonio Zanardo <zanardo@gmail.com>
#

from bottle import route, view, get, post, redirect, \
    response, request, static_file, abort
from functools import wraps
from hashlib import sha1
from uuid import uuid4

import os
import re
import yaml
import zlib
import random
import bottle
import os.path
import datetime
import subprocess
import zpgdb as db

from clogd import __VERSION__
from clogd.log import log

# Default configuration. The environment variable CLOGD_CONF should be
# declared before importing this module. This envvar should point to a yaml
# file containing a dictionary with keys and values that override the default
# configuration.
CONFIG = {
    'pg_host': '127.0.0.1',
    'pg_port': 5432,
    'pg_db': 'clog_dev',
    'pg_user': 'clog_dev',
    'pg_pass': '**password**',
}

bottle.TEMPLATE_PATH.insert(
    0, os.path.join(os.path.dirname(__file__), 'views')
)

# Bottle has a low limit for post data. Let's make it larger.
bottle.BaseRequest.MEMFILE_MAX = 10 * 1024 * 1024


def duration_to_human(seconds):
    ms = (seconds - int(seconds)) * 100
    s = seconds
    m = seconds/60
    h = seconds/3600
    return '%02d:%02d:%02d.%03d' % (h, m, s, ms)


def date_to_human(dt):
    if dt is None:
        return ''
    return dt.strftime('%Y-%m-%d')


def getctx():
    user = currentuser()
    user_is_admin = userisadmin(user)
    return dict(version=__VERSION__, username=user, isadmin=user_is_admin)


def requires_auth(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        session_id = request.get_cookie('clog')
        if not session_id or not validatesession(session_id):
            return redirect('/login')
        return f(*args, **kwargs)
    return decorated


def requires_admin(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        session_id = request.get_cookie('clog')
        if not session_id or not validatesession(session_id) or \
                not userisadmin(currentuser()):
            return 'not authorized'
        return f(*args, **kwargs)
    return decorated


def validateuserdb(user, passwd):
    passwdsha1 = sha1(passwd).hexdigest()
    with db.trans() as c:
        c.execute('''
            SELECT username
            FROM users
            WHERE username = %(user)s
              AND password = %(passwdsha1)s
        ''', locals())
        r = c.fetchone()
        return bool(r)


def validatesession(session_id):
    with db.trans() as c:
        c.execute('''
            SELECT session_id
            FROM sessions
            WHERE session_id = %(session_id)s
        ''', locals())
        r = c.fetchone()
        return bool(r)


def currentuser():
    session_id = request.get_cookie('clog')
    with db.trans() as c:
        c.execute('''
            SELECT username
            FROM sessions
            WHERE session_id = %(session_id)s
        ''', locals())
        return c.fetchone()['username']


def userisadmin(username):
    with db.trans() as c:
        c.execute('''
            SELECT is_admin
            FROM users
            WHERE username = %(username)s
        ''', locals())
        return c.fetchone()['is_admin']


def removesession(session_id):
    with db.trans() as c:
        c.execute('''
            DELETE FROM sessions
            WHERE session_id = %(session_id)s
        ''', locals())


def makesession(user):
    with db.trans() as c:
        session_id = str(uuid4())
        c.execute('''
            INSERT INTO sessions
                (session_id, username)
            VALUES
                (%(session_id)s, %(user)s)
        ''', locals())
        return session_id


def get_job_id(computername, computeruser, script):
    with db.trans() as c:
        c.execute('''
            SELECT id
            FROM jobs
            WHERE computername = %(computername)s
                AND computeruser = %(computeruser)s
                AND script = %(script)s
        ''', locals())
        r = c.fetchone()
        if not r:
            return None
        else:
            return r[0]


@get('/admin')
@view('admin')
@requires_auth
@requires_admin
def admin():
    users = []
    with db.trans() as c:
        c.execute('''
            SELECT username, is_admin
            FROM users
            ORDER BY username
        ''')
        for user in c:
            user = dict(user)
            users.append(user)
    return dict(ctx=getctx(), users=users)


@get('/admin/remove-user/:username')
@requires_auth
@requires_admin
def removeuser(username):
    if username == currentuser():
        return 'cant remove current user!'
    with db.trans() as c:
        c.execute('''
            DELETE FROM sessions
            WHERE username = %(username)s
        ''', locals())
        c.execute('''
            DELETE FROM users
            WHERE username = %(username)s
        ''', locals())
    return redirect('/admin')


@post('/admin/save-new-user')
@requires_auth
@requires_admin
def newuser():
    username = request.forms.username
    if username.strip() == '':
        return 'invalid user!'
    password = str(int(random.random() * 999999))
    sha1password = sha1(password).hexdigest()
    with db.trans() as c:
        c.execute('''
            INSERT INTO users
                (username, password, is_admin)
            VALUES
                (%(username)s, %(sha1password)s, 'f')
        ''', locals())
    return u'user %s created with password %s' % (username, password)


@get('/admin/force-new-password/:username')
@requires_auth
@requires_admin
def forceuserpassword(username):
    password = str(int(random.random() * 999999))
    sha1password = sha1(password).hexdigest()
    if username == currentuser():
        return 'cant change password for current user!'
    with db.trans() as c:
        c.execute('''
            UPDATE users
            SET password = %(sha1password)s
            WHERE username = %(username)s
        ''', locals())
    return u'user %s had password changed to: %s' % (username, password)


@get('/admin/change-user-admin-status/:username/:status')
@requires_auth
@requires_admin
def changeuseradminstatus(username, status):
    if username == currentuser():
        return 'cant change admin status for current user!'
    if status not in ('0', '1'):
        abort(400, "invalid status")
    status = bool(int(status))
    with db.trans() as c:
        c.execute('''
            UPDATE users
            SET is_admin = %(status)s
            WHERE username = %(username)s
        ''', locals())
    return redirect('/admin')


@get('/login')
@view('login')
def login():
    return dict(version=__VERSION__)


@post('/login')
def validatelogin():
    user = request.forms.user
    passwd = request.forms.passwd
    if validateuserdb(user, passwd):
        session_id = makesession(user)
        response.set_cookie('clog', session_id)
        return redirect('/')
    else:
        return 'invalid user or password'


@get('/logout')
def logout():
    session_id = request.get_cookie('clog')
    if session_id:
        removesession(session_id)
        response.delete_cookie('clog')
    return redirect('/login')


@get('/change-password')
@view('change-password')
@requires_auth
def changepassword():
    return dict(ctx=getctx())


@post('/change-password')
@requires_auth
def changepasswordsave():
    oldpasswd = request.forms.oldpasswd
    newpasswd = request.forms.newpasswd
    newpasswd2 = request.forms.newpasswd2
    username = currentuser()
    if not validateuserdb(username, oldpasswd):
        return 'invalid current password!'
    if newpasswd.strip() == '' or newpasswd2.strip() == '':
        return 'invalid new password!'
    if newpasswd != newpasswd2:
        return 'new passwords do not match!'
    passwdsha1 = sha1(newpasswd).hexdigest()
    with db.trans() as c:
        c.execute('''
            UPDATE users
            SET password = %(passwdsha1)s
            WHERE username = %(username)s
        ''', locals())
    return redirect('/')


@route('/static/:filename')
def static(filename):
    if not re.match(r'^[\w\d\-]+\.[\w\d\-]+$', filename):
        abort(400, "invalid filename")
    root = os.path.dirname(__file__)
    return static_file('static/%s' % filename, root=root)


@get('/jobs/<computername>/<computeruser>/<script>/<id>')
@requires_auth
def joboutput(computername, computeruser, script, id):
    if not re.match(r'^[a-f0-9-]{36}$', id):
        raise ValueError('invalid id')
    output = ''
    with db.trans() as c:
        c.execute('''
            SELECT o.output
            FROM jobhistory AS h
                INNER JOIN jobs AS j
                    ON h.job_id = j.id
                INNER JOIN outputs AS o
                    ON o.sha1 = h.output_sha1
            WHERE j.computername = %(computername)s
                AND j.computeruser = %(computeruser)s
                AND j.script = %(script)s
                AND h.id = %(id)s
        ''', locals())
        r = c.fetchone()
        if not r:
            response.status = 404
            return 'not found'
        else:
            response.content_type = 'text/plain; charset=utf-8'
            return zlib.decompress(r['output'])


@get('/jobs/<computername>/<computeruser>/<script>/')
@view('history')
@requires_auth
def jobhistory(computername, computeruser, script):
    ctx = getctx()
    ctx['computername'] = computername
    ctx['computeruser'] = computeruser
    ctx['script'] = script
    with db.trans() as c:
        c.execute('''
            SELECT h.id
                , j.computername
                , j.computeruser
                , j.script
                , j.datestarted
                , h.datefinished
                , h.status
                , h.duration
            FROM jobhistory AS h
            INNER JOIN jobs AS j
                ON j.id = h.job_id
            WHERE j.computername = %(computername)s
                AND j.computeruser = %(computeruser)s
                AND j.script = %(script)s
            ORDER BY j.computername
                , j.computeruser
                , j.script
                , h.datestarted DESC
        ''', locals())
        history = []
        for hist in c:
            h = dict(hist)
            h['duration'] = duration_to_human(h['duration'])
            history.append(h)
        return dict(history=history, ctx=ctx)


@get('/history')
@view('historytable')
@requires_auth
def allhistory():
    offset = 0
    if 'offset' in request.query:
        if re.match(r'^\d+$', request.query.offset):
            offset = int(request.query.offset)*25
    with db.trans() as c:
        c.execute('''
            SELECT h.id
                , j.computername
                , j.computeruser
                , j.script
                , h.datestarted
                , h.datefinished
                , h.status
                , h.duration
            FROM jobhistory AS h
                INNER JOIN jobs AS j
                    ON j.id = h.job_id
            ORDER BY h.datestarted DESC LIMIT 25 OFFSET %(offset)s
        ''', locals())
        history = []
        for hist in c:
            h = dict(hist)
            h['duration'] = duration_to_human(h['duration'])
            history.append(h)
        return dict(history=history, offset=offset)


@get('/config-job/<computername>/<computeruser>/<script>/')
@view('config-job')
@requires_auth
@requires_admin
def configjob(computername, computeruser, script):
    ctx = getctx()
    ctx['computername'] = computername
    ctx['computeruser'] = computeruser
    ctx['script'] = script
    daystokeep = 30
    with db.trans() as c:
        c.execute('''
            SELECT c.daystokeep
            FROM jobconfig AS c
                INNER JOIN jobs AS j
                    ON j.id = c.job_id
            WHERE j.computername = %(computername)s
                AND j.computeruser = %(computeruser)s
                AND j.script = %(script)s
        ''', locals())
        r = c.fetchone()
        if r:
            daystokeep = r['daystokeep']
        c.execute('''
            SELECT a.email
            FROM jobconfigalert AS a
                INNER JOIN jobs AS j
                    ON j.id = a.job_id
            WHERE j.computername = %(computername)s
                AND j.computeruser = %(computeruser)s
                AND j.script = %(script)s
            ''', locals())
        emails = []
        for r in c:
            emails.append(r['email'])
        return dict(ctx=ctx, daystokeep=daystokeep, emails="\n".join(emails))


@post('/purge-job/<computername>/<computeruser>/<script>/')
@requires_auth
@requires_admin
def purgejob(computername, computeruser, script):
    job_id = get_job_id(computername, computeruser, script)
    with db.trans() as c:
        c.execute('''
            DELETE FROM jobhistory
            WHERE job_id = %(job_id)s
        ''', locals())
        c.execute('''
            DELETE FROM jobconfig
            WHERE job_id = %(job_id)s
        ''', locals())
        c.execute('''
            DELETE FROM jobconfigalert
            WHERE job_id = %(job_id)s
        ''', locals())
        c.execute('''
            DELETE FROM jobs
            WHERE id = %(job_id)s
        ''', locals())
    return redirect('/')


@post('/save-daystokeep/<computername>/<computeruser>/<script>/')
@requires_auth
@requires_admin
def savedaystokeep(computername, computeruser, script):
    daystokeep = request.forms.daystokeep
    if not re.match(r'^\d+$', daystokeep):
        abort(400, "invalid days to keep")
    daystokeep = int(daystokeep)
    if daystokeep < 0:
        return 'days to keep must be >= 0'
    job_id = get_job_id(computername, computeruser, script)
    with db.trans() as c:
        c.execute('''
            UPDATE jobconfig
            SET daystokeep = %(daystokeep)s
            WHERE job_id = %(job_id)s
        ''', locals())
        if c.rowcount == 0:
            c.execute('''
                INSERT INTO jobconfig
                    (job_id, daystokeep)
                VALUES
                    (%(job_id)s, %(daystokeep)s)
            ''', locals())
    return redirect('/config-job/' + computername + '/' +
                    computeruser + '/' + script + '/')


@post('/save-alertemails/<computername>/<computeruser>/<script>/')
@requires_auth
@requires_admin
def savealertemails(computername, computeruser, script):
    job_id = get_job_id(computername, computeruser, script)
    with db.trans() as c:
        c.execute('''
            DELETE FROM jobconfigalert
            WHERE job_id = %(job_id)s
        ''', locals())
        for email in request.forms.emails.split():
            c.execute('''
                INSERT INTO jobconfigalert
                VALUES
                    (%(job_id)s, %(email)s)
            ''', locals())
    return redirect('/config-job/' + computername + '/' +
                    computeruser + '/' + script + '/')


@get('/')
@view('jobs')
@requires_auth
def index():
    return dict(ctx=getctx())


@get('/jobs')
@view('jobstable')
@requires_auth
def jobs():
    with db.trans() as c:
        c.execute('''
            SELECT computername
                , computeruser
                , script
                , date_last_success
                , date_last_failure
                , last_status
                , last_duration
            FROM jobs
            ORDER BY computername
                , computeruser
                , script
        ''')
        jobs = []
        for job in c:
            j = dict(job)
            j['date_last_success'] = date_to_human(j['date_last_success'])
            j['date_last_failure'] = date_to_human(j['date_last_failure'])
            j['last_duration'] = duration_to_human(j['last_duration'])
            jobs.append(j)
        return dict(jobs=jobs)


@post('/')
def newjob():
    rid_regexp = r'^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$'
    rid = request.forms.id
    if not re.match(rid_regexp, rid):
        abort(400, "invalid job id")

    start_time = request.forms.start_time
    if not re.match(r'^[\d+\.]+$', start_time):
        abort(400, "invalid start time")
    start_time = datetime.datetime.fromtimestamp(int(float(start_time)))

    end_time = request.forms.end_time
    if not re.match(r'^[\d+\.]+$', end_time):
        abort(400, "invalid end time")
    end_time = datetime.datetime.fromtimestamp(int(float(end_time)))

    duration = request.forms.duration
    if not re.match(r'^[0-9\.]+$', duration):
        abort(400, "invalid duration")

    status = request.forms.status
    if status not in ('fail', 'ok'):
        abort(400, "invalid status")

    script = request.forms.script
    if not re.match(r'^[a-zA-Z0-9\-_\.]+$', script):
        abort(400, "invalid script name")

    output = request.forms.output.encode('utf-8') or ''
    outputz = buffer(zlib.compress(output))

    computername = request.forms.computername
    computeruser = request.forms.username

    # Windows
    if '\\' in computeruser:
        computeruser = computeruser.split('\\')[-1]

    ip = request.remote_addr

    log.info('new job status from %s@%s/%s (%s)', computeruser, computername,
             script, ip)
    log.info('  id: %s', rid)

    output_sha1 = sha1(output).hexdigest()
    with db.trans() as c:
        try:
            c.execute('''
                INSERT INTO outputs
                    (sha1, output)
                VALUES
                    (%(output_sha1)s, %(outputz)s)
            ''', locals())
        except db.psycopg2.IntegrityError:
            pass

    job_id = get_job_id(computername, computeruser, script)
    if not job_id:
        with db.trans() as c:
            c.execute('''
                INSERT INTO jobs (
                    computername, computeruser, script,
                    last_status, last_duration
                )
                VALUES (
                    %(computername)s, %(computeruser)s, %(script)s,
                    %(status)s, %(duration)s
                )
                RETURNING id
            ''', locals())
            job_id = c.fetchone()[0]
    try:
        with db.trans() as c:
            c.execute('''
                INSERT INTO jobhistory (
                    id, job_id, ip, datestarted, datefinished,
                    status, duration, output_sha1
                )
                VALUES (
                    %(rid)s, %(job_id)s, %(ip)s, %(start_time)s
                    %(end_time)s, %(status)s, %(duration)s, %(output_sha1)s
                )
                ''', locals())
            if status == 'ok':
                c.execute('''
                    UPDATE jobs
                    SET date_last_success = %(start_time)s
                        , last_status = 'ok'
                        , last_duration = %(duration)s
                    WHERE id = %(job_id)s
                ''', locals())
            else:
                c.execute('''
                    UPDATE jobs
                    SET date_last_failure = %(start_time)s
                        , last_status = 'fail'
                        , last_duration = %(duration)s
                    WHERE id = %(job_id)s
                ''', locals())
    except db.psycopg2.IntegrityError:
        # Ignoring duplicate insertion.
        return 'ok'
    else:
        emails = getalertemails(computername, computeruser, script)
        if emails:
            if status == 'fail':
                for email in emails:
                    log.info("  job failed, sending alert to %s", email)
                    send_alert(
                        email, computername, computeruser, script,
                        status, output)
            elif status == 'ok':
                with db.trans() as c:
                    c.execute('''
                        SELECT status
                        FROM jobhistory
                        WHERE job_id = %(job_id)s
                        ORDER BY datestarted DESC LIMIT 1 OFFSET 1
                    ''', locals())
                    r = c.fetchone()
                    if r and r['status'] == 'fail':
                        for email in emails:
                            log.info("  job ok, sending alert to %s", email)
                            send_alert(
                                email, computername, computeruser, script,
                                status, output
                            )
        return 'ok'


# Get notification e-mails for a job.
def getalertemails(computername, computeruser, script):
    job_id = get_job_id(computername, computeruser, script)
    with db.trans() as c:
        c.execute('''
            SELECT email
            FROM jobconfigalert
            WHERE job_id = %(job_id)s
        ''', locals())
        emails = []
        for row in c:
            emails.append(row['email'])
        return emails


# Delete login sessions older than 7 days
def purge_sessions():
    with db.trans() as c:
        c.execute('''
            DELETE FROM sessions
            WHERE date(now()) - date(date_login) > 7
        ''')
        if c.rowcount > 0:
            log.info('purged %s login sessions', c.rowcount)


# Delete old entries on jobhistory from database.
def purge_jobhistory():
    with db.trans() as c:
        c.execute('''
            SELECT id
            FROM jobs
        ''')
        for job in c:
            job_id = job['id']
            with db.trans() as c2:
                c2.execute('''
                    SELECT daystokeep
                    FROM jobconfig
                    WHERE job_id = %(job_id)s
                ''', locals())
                daystokeep = 30
                r = c2.fetchone()
                if r:
                    daystokeep = r['daystokeep']
                with db.trans() as c3:
                    c3.execute('''
                        DELETE FROM jobhistory
                        WHERE date(now()) - date(datestarted) > %(daystokeep)s
                            AND job_id = %(job_id)s
                    ''', locals())
                    if c3.rowcount > 0:
                        log.debug("purged %s entries for jobhistory",
                                  c3.rowcount)


# Delete unreferenced entries from outputs.
def purge_outputs():
    with db.trans() as c:
        c.execute('''
            DELETE FROM outputs
            WHERE sha1 NOT IN (
                SELECT DISTINCT output_sha1 FROM jobhistory
            )
        ''')
        if c.rowcount > 0:
            log.debug("purged %s entries for outputs", c.rowcount)


def send_alert(email, computername, computeruser, script, status, output):
    subject = ''
    body = ''
    if status == 'fail':
        subject = 'clog: job {} failed for {}@{}'.format(
            script,
            computeruser,
            computername
        )
        body = output
    elif status == 'ok':
        subject = 'clog: job {} back to normal for {}@{}'.format(
            script,
            computeruser,
            computername
        )
    body += '\n\nThis is an automatic notification sent by ' + \
            'clog (https://github.com/zanardo/clog)'
    s = subprocess.Popen(['mail', '-s', subject, email], stdin=subprocess.PIPE)
    s.communicate(body)


# Purge expired data.
@get('/cleanup')
def cleanup():
    log.info('starting maintenance')
    purge_jobhistory()
    purge_outputs()
    purge_sessions()
    log.info('finishing maintenance')


app = bottle.default_app()

# Reading configuration file.
if 'CLOGD_CONF' in os.environ:
    log.info('reading configuration from %s', os.environ['CLOGD_CONF'])
    with open(os.environ['CLOGD_CONF'], 'r') as fp:
        conf = yaml.load(fp)
        for k in CONFIG:
            if k in conf:
                CONFIG[k] = conf[k]

db.config_connection(
    CONFIG['pg_host'], CONFIG['pg_port'], CONFIG['pg_user'],
    CONFIG['pg_pass'], CONFIG['pg_db'])
