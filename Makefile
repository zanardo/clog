PYTHON=python2

all: dev

dev:
	test -d .venv || virtualenv -p $(PYTHON) .venv
	.venv/bin/pip install -r requirements.txt

run-server-devel: venv
	while :; do CLOGD_CONF="`pwd`/config.dev.yml" .venv/bin/python ./run-server-devel.py; sleep 2; done

install-cli:
	sudo install -o root -g root -m 755 scripts/clog /usr/local/bin/clog

run-server: venv
	CLOGD_CONF=`pwd`/config.yml .venv/bin/waitress-serve --host 0.0.0.0 --port 27890 clogd:app

clean:
	rm -rf .venv


.PHONY: install-cli run-server run-server-devel
