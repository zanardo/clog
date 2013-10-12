all: db venv

db:
	@test -f clog.db || ( sqlite3 clog.db < schema.sql \
		&& echo "creating clog.db" )
	@test -d clog.db.logs || ( mkdir clog.db.logs \
		&& chmod 700 clog.db.logs && echo "creating clog.db.logs" )

venv: .venv/bin/activate

.venv/bin/activate: requirements.txt
	test -d .venv || virtualenv --no-site-packages --distribute .venv
	. .venv/bin/activate; pip install -r requirements.txt
	touch .venv/bin/activate

install-cli:
	sudo install -o root -g root -m 755 clog /usr/bin/clog
