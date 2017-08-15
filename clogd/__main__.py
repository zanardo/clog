#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from bottle import run
from clogd import app
from os import environ

if 'CLOGD_CONF' not in environ:
    environ['CLOGD_CONF'] = 'config.dev.yml'

run(
    app=app,
    host='127.0.0.1',
    port=7890,
    interval=1,
    reloader=True,
    debug=True,
)
