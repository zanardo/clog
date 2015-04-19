all: venv

venv: .venv/bin/activate

.venv/bin/activate: requirements.txt
	test -d .venv || virtualenv-2.7 --no-site-packages --distribute .venv
	. .venv/bin/activate; pip install -r requirements.txt
	touch .venv/bin/activate

install-cli:
	sudo install -o root -g root -m 755 clog /usr/bin/clog

clean:
	rm -f *.pyc
	rm -rf .venv


.PHONY: all venv install-cli clean
