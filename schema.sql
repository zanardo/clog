CREATE TABLE jobs (
	id integer not null primary key,
	computername text not null,
	computeruser text not null,
	script text not null,
	date_last_success timestamp,
	date_last_failure timestamp,
	last_status text,
	last_duration float
);
CREATE UNIQUE INDEX idx_jobs_job ON jobs(computername, computeruser, script);

CREATE TABLE jobhistory (
	id text not null primary key,
	job_id integer not null references jobs(id),
	ip text not null,
	datestarted timestamp not null,
	datefinished timestamp not null,
	duration float not null,
	status text not null,
	output blob
);

CREATE TABLE jobconfig (
	job_id integer not null references jobs(id) primary key,
	daystokeep int not null
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

CREATE TABLE jobconfigalert (
	job_id integer not null references jobs(id),
	email text not null,
	PRIMARY KEY(job_id, email)
);

INSERT INTO users ( username, password, is_admin )
VALUES ( 'admin', 'd033e22ae348aeb5660fc2140aec35850c4da997', 1 );
