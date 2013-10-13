all: db venv

db:
	@test -f clogd.db || ( sqlite3 clogd.db < schema.sql \
		&& echo "creating clogd.db" )
	@test -d clogd.db.logs || ( mkdir clogd.db.logs \
		&& chmod 700 clogd.db.logs && echo "creating clogd.db.logs" )

venv: .venv/bin/activate

.venv/bin/activate: requirements.txt
	test -d .venv || virtualenv --no-site-packages --distribute .venv
	. .venv/bin/activate; pip install -r requirements.txt
	touch .venv/bin/activate

install-cli:
	sudo install -o root -g root -m 755 clog /usr/bin/clog
