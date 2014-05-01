--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: clog; Type: DATABASE; Schema: -; Owner: clog
--

CREATE DATABASE clog WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'en_US.UTF-8' LC_CTYPE = 'en_US.UTF-8';


ALTER DATABASE clog OWNER TO clog;

\connect clog

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: jobconfig; Type: TABLE; Schema: public; Owner: clog; Tablespace: 
--

CREATE TABLE jobconfig (
    job_id integer NOT NULL,
    daystokeep integer NOT NULL
);


ALTER TABLE public.jobconfig OWNER TO clog;

--
-- Name: jobconfigalert; Type: TABLE; Schema: public; Owner: clog; Tablespace: 
--

CREATE TABLE jobconfigalert (
    job_id integer NOT NULL,
    email text NOT NULL
);


ALTER TABLE public.jobconfigalert OWNER TO clog;

--
-- Name: jobhistory; Type: TABLE; Schema: public; Owner: clog; Tablespace: 
--

CREATE TABLE jobhistory (
    id text NOT NULL,
    job_id integer NOT NULL,
    ip text NOT NULL,
    datestarted timestamp(0) without time zone NOT NULL,
    datefinished timestamp(0) without time zone NOT NULL,
    duration real NOT NULL,
    status text NOT NULL,
    output_sha1 character(40) NOT NULL
);


ALTER TABLE public.jobhistory OWNER TO clog;

--
-- Name: jobs; Type: TABLE; Schema: public; Owner: clog; Tablespace: 
--

CREATE TABLE jobs (
    id integer NOT NULL,
    computername text NOT NULL,
    computeruser text NOT NULL,
    script text NOT NULL,
    date_last_success timestamp(0) without time zone,
    date_last_failure timestamp(0) without time zone,
    last_status text,
    last_duration real
);


ALTER TABLE public.jobs OWNER TO clog;

--
-- Name: jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: clog
--

CREATE SEQUENCE jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.jobs_id_seq OWNER TO clog;

--
-- Name: jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: clog
--

ALTER SEQUENCE jobs_id_seq OWNED BY jobs.id;


--
-- Name: outputs; Type: TABLE; Schema: public; Owner: clog; Tablespace: 
--

CREATE TABLE outputs (
    sha1 character(40) NOT NULL,
    output bytea NOT NULL
);


ALTER TABLE public.outputs OWNER TO clog;

--
-- Name: sessions; Type: TABLE; Schema: public; Owner: clog; Tablespace: 
--

CREATE TABLE sessions (
    session_id text NOT NULL,
    date_login timestamp(0) without time zone DEFAULT now() NOT NULL,
    username text NOT NULL
);


ALTER TABLE public.sessions OWNER TO clog;

--
-- Name: users; Type: TABLE; Schema: public; Owner: clog; Tablespace: 
--

CREATE TABLE users (
    username text NOT NULL,
    password text NOT NULL,
    is_admin boolean DEFAULT false NOT NULL
);


ALTER TABLE public.users OWNER TO clog;

--
-- Name: id; Type: DEFAULT; Schema: public; Owner: clog
--

ALTER TABLE ONLY jobs ALTER COLUMN id SET DEFAULT nextval('jobs_id_seq'::regclass);


--
-- Data for Name: jobconfig; Type: TABLE DATA; Schema: public; Owner: clog
--

COPY jobconfig (job_id, daystokeep) FROM stdin;
\.


--
-- Data for Name: jobconfigalert; Type: TABLE DATA; Schema: public; Owner: clog
--

COPY jobconfigalert (job_id, email) FROM stdin;
\.


--
-- Data for Name: jobhistory; Type: TABLE DATA; Schema: public; Owner: clog
--

COPY jobhistory (id, job_id, ip, datestarted, datefinished, duration, status, output) FROM stdin;
\.


--
-- Data for Name: jobs; Type: TABLE DATA; Schema: public; Owner: clog
--

COPY jobs (id, computername, computeruser, script, date_last_success, date_last_failure, last_status, last_duration) FROM stdin;
\.


--
-- Name: jobs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: clog
--

SELECT pg_catalog.setval('jobs_id_seq', 1, false);


--
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: clog
--

COPY sessions (session_id, date_login, username) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: clog
--

COPY users (username, password, is_admin) FROM stdin;
admin	d033e22ae348aeb5660fc2140aec35850c4da997	t
\.


--
-- Name: jobconfig_pkey; Type: CONSTRAINT; Schema: public; Owner: clog; Tablespace: 
--

ALTER TABLE ONLY jobconfig
    ADD CONSTRAINT jobconfig_pkey PRIMARY KEY (job_id);


--
-- Name: jobconfigalert_pkey; Type: CONSTRAINT; Schema: public; Owner: clog; Tablespace: 
--

ALTER TABLE ONLY jobconfigalert
    ADD CONSTRAINT jobconfigalert_pkey PRIMARY KEY (job_id, email);


--
-- Name: jobhistory_pkey; Type: CONSTRAINT; Schema: public; Owner: clog; Tablespace: 
--

ALTER TABLE ONLY jobhistory
    ADD CONSTRAINT jobhistory_pkey PRIMARY KEY (id);


--
-- Name: jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: clog; Tablespace: 
--

ALTER TABLE ONLY jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (id);


--
-- Name: outputs_pkey; Type: CONSTRAINT; Schema: public; Owner: clog; Tablespace: 
--

ALTER TABLE ONLY outputs
    ADD CONSTRAINT outputs_pkey PRIMARY KEY (sha1);


--
-- Name: sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: clog; Tablespace: 
--

ALTER TABLE ONLY sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (session_id);


--
-- Name: users_pkey; Type: CONSTRAINT; Schema: public; Owner: clog; Tablespace: 
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (username);


--
-- Name: idx_jobhistory_datestarted; Type: INDEX; Schema: public; Owner: clog; Tablespace: 
--

CREATE INDEX idx_jobhistory_datestarted ON jobhistory USING btree (datestarted);


--
-- Name: idx_jobhistory_job_id; Type: INDEX; Schema: public; Owner: clog; Tablespace: 
--

CREATE INDEX idx_jobhistory_job_id ON jobhistory USING btree (job_id);


--
-- Name: idx_jobs_job; Type: INDEX; Schema: public; Owner: clog; Tablespace: 
--

CREATE UNIQUE INDEX idx_jobs_job ON jobs USING btree (computername, computeruser, script);


--
-- Name: jobconfig_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: clog
--

ALTER TABLE ONLY jobconfig
    ADD CONSTRAINT jobconfig_job_id_fkey FOREIGN KEY (job_id) REFERENCES jobs(id);


--
-- Name: jobconfigalert_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: clog
--

ALTER TABLE ONLY jobconfigalert
    ADD CONSTRAINT jobconfigalert_job_id_fkey FOREIGN KEY (job_id) REFERENCES jobs(id);


--
-- Name: jobhistory_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: clog
--

ALTER TABLE ONLY jobhistory
    ADD CONSTRAINT jobhistory_job_id_fkey FOREIGN KEY (job_id) REFERENCES jobs(id);


--
-- Name: jobhistory_output_sha1_fkey; Type: FK CONSTRAINT; Schema: public; Owner: clog
--

ALTER TABLE ONLY jobhistory
    ADD CONSTRAINT jobhistory_output_sha1_fkey FOREIGN KEY (output_sha1) REFERENCES outputs(sha1);


--
-- Name: sessions_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: clog
--

ALTER TABLE ONLY sessions
    ADD CONSTRAINT sessions_username_fkey FOREIGN KEY (username) REFERENCES users(username);


--
-- PostgreSQL database dump complete
--

