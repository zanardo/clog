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