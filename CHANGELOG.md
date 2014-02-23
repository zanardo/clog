clog changes
============

## 0.5 - (in development)

## v0.4.1 - 2013-02-22

- Fix last duration field not being saved.

## v0.4 - 2014-02-15

- Human readable job duration.
- Human readable last success and last failure date on index page.
- Reestructured database for better performance on jobs and history pages.
- Database schema auto migration on clog server startup.

## v0.3 - 2013-11-15

- Rewrite clog client in Go, for easier deployment and fast startup time.

## v0.2 - 2013-11-06

- E-Mail alerts when job fails and when job backs to normal after a failure.
- Per job configuration of maximum number of days to maintain history entries.
- New background maintenance thread to remove expired history entries.
- Web page styling changes.
- All jobs history pagination.

## v0.1.1 - 2013-10-25

- Upper HTTP POST size limit to 10MB.

## v0.1 - 2013-10-20

- Initial version.
