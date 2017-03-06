run-server-devel: venv
	while :; do CLOGD_CONF="`pwd`/config.dev.yml" .venv/bin/python ./run-server-devel.py; sleep 2; done

install-cli:
	sudo install -o root -g root -m 755 scripts/clog /usr/local/bin/clog

run-server: venv
	CLOGD_CONF=`pwd`/config.yml .venv/bin/waitress-serve --host 0.0.0.0 --port 27890 clogd:app


.PHONY: install-cli run-server run-server-devel
