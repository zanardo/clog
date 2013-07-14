all: db venv

db:
	@test -f clog.db || ( sqlite3 clog.db < schema.sql \
		&& echo "creating clog.db" )

venv: .venv/bin/activate

.venv/bin/activate: requirements.txt
	test -d .venv || virtualenv --no-site-packages --distribute .venv
	. .venv/bin/activate; pip install -r requirements.txt
	touch .venv/bin/activate
