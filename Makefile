all: db venv

venv: .venv/bin/activate

.venv/bin/activate: requirements.txt
	test -d .venv || virtualenv --no-site-packages --distribute .venv
	. .venv/bin/activate; pip install -r requirements.txt
	touch .venv/bin/activate

install-cli:
	sudo install -o root -g root -m 755 clog/clog /usr/bin/clog

run-server-devel: venv
	while :; do ./.venv/bin/python clogd config_devel.yaml ; sleep 0.5 ; done

run-server: venv
	./.venv/bin/python clogd config.yaml
