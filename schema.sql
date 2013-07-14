CREATE TABLE jobhistory (
	id text not null primary key,
	script text not null,
	computername text not null,
	ip text not null,
	computeruser text not null,
	datestarted timestamp not null,
	datefinished timestamp not null,
	status integer not null
);