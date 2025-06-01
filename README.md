# squanchy
> Single executable rest authentication api that doesn't fail.

The perfect tool to use just to "try an **idea** out" before handling authentication in a more scalable manner!

## Features

<img align="right" width="215" src="./squanchy.png">

* It will always spin up a **local** authentication API, no matter what!
* Extremely configurable with flags and environment variables.
* Reliable using persistent and backward compatible migrations on sqlite3.
* As performant as the chi router allows to handle requests.
* Structured JSON responses and beautiful simple logging.
* Documented and understandable which truly isn't a given...
* Security features and extensively **audited** for SQL injections, BOLA and more!

## Usage
```
you@computer$ squanchy
```
```
Listening to requests on :6900/api
```

Use **flags**(prioritized) and/or **environment variables**(find them in .env.example) to change the program's behaviour!

### Flags
```
you@computer$ squanchy -h
```
```
Usage: squanchy [--port] [--prefix] [--rate-limit] [--extra-rate-limit] [--pepper] [--bcrypt-cost] [--lock-time] [--attempts] [--jwt-secret] [--jwt-expiration] [--database] [--journal-mode] [--totp] [--totp-skew] [--totp-algorithm] [--smtp] [--smtp-address] [--smtp-port] [--smtp-user] [--smtp-password] [--smtp-from] [--origins] [--easter-egg] [--version]

Options:
  --port, -p             specifies the port to bind to [default: 6900, env: PORT]
  --prefix               speicifes a prefix [default: /api, env: PREFIX]
  --soft-rate-limit, -r  specifies the soft rate limit(1 hour window) [default: 60, env: RATE_LIMIT]
  --hard-rate-limit, -e  specifies the hard rate limit for sensitive endpoints(1 hour window) [default: 10, env: EXTRA_RATE_LIMIT]
  --pepper               specifies a password pepper [env: PEPPER]
  --bcrypt-cost          specifies the bcrypt cost [default: 10, env: COST]
  --lock-time, -l        specifies the lock time after too many failed attempts in hours [default: 6, env: LOCK_TIME]
  --attempts, -a         specifies the maximum number of failed attempts before being locked out [default: 9, env: LOCK_ATTEMPTS]
  --jwt-secret, -c       specifies the JWT secret [default: d3f4ult_jwt$$_secret_cha:)nge_me?, env: JWT_SECRET]
  --jwt-expiration, -x   specifies the expiration time of JWTs in hours [default: 24, env: JWT_EXPIRATION]
  --database, -d         specifies the database [default: ./squanchy.db, env: DATABASE_PATH]
  --journal-mode, -j     specifies the sqlite3 database journal mode [default: DELETE, env: JOURNAL_MODE]
  --totp, -t             specifies whether to enable 2FA using TOTP or not [default: false, env: ENABLE_TOTP]
  --totp-skew            specifies the TOTP validation skew [default: 0, env: TOTP_SKEW]
  --totp-algorithm       specifies the TOTP algorithm to use [default: 1, env: TOTP_ALGORITHM]
  --smtp, -s             specifies whether to enable emails using SMTP [default: false, env: ENABLE_SMTP]
  --smtp-address         specifies the SMTP server address [env: SMTP_ADDRESS]
  --smtp-port            specifies the SMTP server port [default: 25, env: SMTP_PORT]
  --smtp-user            specifies the SMTP server user [env: SMTP_USER]
  --smtp-password        specifies the SMTP server password [env: SMTP_PASSWORD]
  --smtp-from            specifies the SMTP from address [env: SMTP_FROM]
  --origins, -o          specifies the CORS origins as a comma separated list [env: CORS_ORIGINS]
  --easter-egg, -g       specifies whether to register the easter egg endpoint [default: false, env: EASTER_EGG]
  --version, -v          prints squanchy's version and exit
  --help, -h             display this help and exit
```
Yes, there might be a lot in hear, but it's definately understandable!

# Documentation
Other than the code itself please read the reference at /openapi and try endpoints out with your favorite HTTP client.

## Installation
Grab a [release](https://github.com/squanchy) for your architecture and operating system or do it another way.

### Use Go's package manager
```
you@computer$ go install github.com/0xsanchez/squanchy@latest
```

### Compile from source
```
you@computer$ CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' ./cmd/squanchy/
```

## Contributions
Now if you wanna add a great feature, just create an issue to ask me about it before working on it just because it's not assured that I will add it once presented as a PR!

You can check out [TODO](./TODO.md) and these will definitely get accepted after review.
