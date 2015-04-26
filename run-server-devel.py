#!/usr/bin/env python2
# -*- coding: utf-8 -*-

from bottle import run
from clogd import app
from os import environ

if not "CLOGD_CONF" in environ:
    environ["CLOGD_CONF"] = "config.dev.yml"

run(app=app, host="127.0.0.1", port=57890,
    interval=1, reloader=True, debug=True)
