begin;

create table jobs (
  id serial not null primary key,
  computername text not null,
  computeruser text not null,
  script text not null,
  date_last_success timestamp(0) without time zone,
  date_last_failure timestamp(0) without time zone,
  last_status text,
  last_duration real
);
create unique index idx_jobs_job on jobs (computername, computeruser, script);

create table jobconfig (
  job_id integer not null references jobs (id),
  daystokeep integer not null
);

create table jobconfigalert (
  job_id integer not null references jobs (id),
  email text not null,
  primary key (job_id, email)
);

create table outputs (
  sha1 character(40) not null primary key,
  output bytea not null
);

create table jobhistory (
  id text not null primary key,
  job_id integer not null references jobs (id),
  ip text not null,
  datestarted timestamp(0) without time zone not null,
  datefinished timestamp(0) without time zone not null,
  duration real not null,
  status text not null,
  output_sha1 character(40) not null references outputs (sha1)
);
create index idx_jobhistory_datestarted on jobhistory (datestarted);
create index idx_jobhistory_job_id on jobhistory (job_id);

create table users (
  username text not null primary key,
  password text not null,
  is_admin boolean default false not null
);

create table sessions (
  session_id text not null primary key,
  date_login timestamp(0) without time zone default now() not null,
  username text not null references users (username)
);

-- Insert default admin user
insert into users (username, password, is_admin)
values ('admin', 'd033e22ae348aeb5660fc2140aec35850c4da997', 't');

commit;
