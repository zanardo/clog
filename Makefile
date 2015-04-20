all: venv

venv: .venv/bin/activate

.venv/bin/activate: requirements.txt
	test -d .venv || virtualenv-2.7 --no-site-packages --distribute .venv
	. .venv/bin/activate; pip install -r requirements.txt
	touch .venv/bin/activate

install-cli:
	sudo install -o root -g root -m 755 scripts/clog /usr/bin/clog

clean:
	rm -f *.pyc
	rm -rf .venv

run-server-devel: venv
	while :; do CLOGD_CONF="`pwd`/config.dev.json" .venv/bin/waitress-serve --host 127.0.0.1 --port 57890 clogd:app; sleep 2; done

run-server: venv
	CLOGD_CONF=`pwd`/config.json .venv/bin/waitress-serve --host 0.0.0.0 --port 27890 clogd:app


.PHONY: all venv install-cli clean
