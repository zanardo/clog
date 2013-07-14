#!/usr/bin/env python
# -*- coding: utf-8 -*-

import sys
import os
import os.path
import subprocess
import uuid
import time
import json

def gethome():
	return os.path.expanduser('~')

def getscriptspath():
	return os.path.join(gethome(), '.clog-scripts')

def getqueuepath():
	return os.path.join(gethome(), '.clog-queue')

def createpaths():
	paths = [ getscriptspath(), getqueuepath() ]
	for p in paths:
		if not os.path.isdir(p):
			print("creating {}...").format(p)
			os.mkdir(p, 0700)

def runqueue():
	print("starting queue dispatch...")
	qp = getqueuepath()
	os.chdir(qp)
	for f in os.listdir('.'):
		if len(f) == 41 and f.endswith('.meta'):
			rid = f[:36]
			print("sending job {}...").format(rid)
			os.unlink("{}.log".format(rid))
			os.unlink("{}.meta".format(rid))
	print("finished queue dispatch")

def runscript(script):
	print("running script {}").format(script)
	sp = os.path.join(getscriptspath(), script)
	if not os.path.isfile(sp):
		print("error: file {} not found!").format(sp)
		sys.exit(1)
	rid = str(uuid.uuid4())
	qlp = os.path.join(getqueuepath(), "{}.log".format(rid))
	with open(qlp, "w") as fp:
		st = time.time()
		ret = subprocess.call(sp, stdout=fp, stderr=subprocess.STDOUT)
		et = time.time()
	qcp = os.path.join(getqueuepath(), rid)
	with open(qcp + ".tmp", "w") as fp:
		json.dump(dict(start_time=st, end_time=et,
			status='OK' if ret == 0 else 'FAIL'), fp)
	os.rename(qcp + '.tmp', qcp + '.meta')
	print("finished with id={}").format(rid)
	return rid

if __name__ == '__main__':

	createpaths()

	if len(sys.argv) == 1:
		runqueue()
	else:
		unit = sys.argv[1]
		rid = runscript(unit)
