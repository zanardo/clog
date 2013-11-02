CREATE TABLE jobhistory (
	id text not null primary key,
	script text not null,
	computername text not null,
	ip text not null,
	computeruser text not null,
	datestarted timestamp not null,
	datefinished timestamp not null,
	duration float not null,
	status text not null
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

CREATE TABLE jobconfig (
	computername text not null,
	computeruser text not null,
	script not null,
	daystokeep int not null,
	PRIMARY KEY(computername, computeruser, script)
);

CREATE TABLE jobconfigalert (
	computername text not null,
	computeruser text not null,
	script text not null,
	email text not null,
	PRIMARY KEY(computername, computeruser, script, email)
);

INSERT INTO users ( username, password, is_admin )
VALUES ( 'admin', 'd033e22ae348aeb5660fc2140aec35850c4da997', 1 );