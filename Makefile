PYTHON=python2

all: dev

dev:
	test -d .venv || virtualenv -p $(PYTHON) .venv
	test -d *.egg-info || .venv/bin/pip install -e .

run-dev: dev
	while :; do CLOGD_CONF="`pwd`/config.dev.yml" \
		.venv/bin/python -m clogd ; \
		sleep 1 ; \
	done

install-cli:
	sudo install -o root -g root -m 755 scripts/clog /usr/local/bin/clog

clean:
	rm -rf .venv *.egg-info


.PHONY: all dev run-dev install-cli clean
