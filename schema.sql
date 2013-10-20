CREATE TABLE jobhistory (
	id text not null primary key,
	script text not null,
	computername text not null,
	ip text not null,
	computeruser text not null,
	datestarted decimal not null,
	datefinished decimal not null,
	status integer not null
);

CREATE TABLE users (
	username text NOT NULL PRIMARY KEY,
	password text NOT NULL,
	is_admin int
);

CREATE TABLE sessions (
	session_id text NOT NULL PRIMARY KEY,
	date_login timestamp NOT NULL DEFAULT ( datetime('now', 'localtime') ),
	username text NOT NULL
);

INSERT INTO users ( username, password, is_admin )
VALUES ( 'admin', 'd033e22ae348aeb5660fc2140aec35850c4da997', 1 );