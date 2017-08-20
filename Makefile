PYTHON=python2

all: .venv

.venv: setup.py
	test -d venv || virtualenv -p $(PYTHON) .venv
	.venv/bin/pip install -U -e .
	@touch .venv

run: .venv
	while :; do CLOGD_CONF="$(CURDIR)/config.dev.yml" \
		.venv/bin/python -m clogd ; \
		sleep 1 ; \
	done

install-cli:
	sudo install -o root -g root -m 755 scripts/clog /usr/local/bin/clog

clean:
	@rm -rf .venv/ *.egg-info/ build/ dist/

.PHONY: all run install-cli clean
