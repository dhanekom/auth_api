
# :warning: Warning - this api is not yet production ready

  

# auth_api

Authentication api written in Go

  

# Description

This api provides you with all the basic features required to manage user registration, authorization (using JSON Web Tokens) and more.

  

## Features

  

- Register users

- Generate verification codes

- Verify users

- Get JWT auth tokens (use in frontend Authorization headers)

- Delete users (admin users only)

  
  

## Setup

- Step 1: Download and install [Go](https://go.dev/doc/install) (requires Go 1.22 or higher)

- Step 2: Clone this git repo

- Step 3: Setup a Postgres DB (e.g using Docker)

- Step 4: Configure environment variables - All configuration is done using environment variables.

The application supports .env files(see a sample .env file below)

```

AUTH_HOST_ADDR=

AUTH_HOST_PORT=80

AUTH_DB_CONNECTION_STRING=postgres://dbname:password@127.0.0.1:5432/auth_db?sslmode=disable

AUTH_JWT_SECRET=supersecretkey

AUTH_VERIFICATION_CODE_LENGTH=6

AUTH_VERIFICATION_MAX_RETRIES=3

AUTH_USER_TOKEN_SECRET=supersecretkey

AUTH_ADMIN_TOKEN_SECRET=supersecretkey

```

- Step 5: Install and start docker - this application uses a [Postgres testcontainer](https://golang.testcontainers.org/modules/postgres/). The docker image will automatically be pulled when you run tests

- Step 6: Build the application

```console

# navigate to the cloned project root directory

go build -o <output path> ./...

```

  

## Testing

[Bruno](https://www.usebruno.com/downloads) is an open source api client for testing APIs. You can install Bruno and open the <code>./testing/bruno/auth_api</code> collection for testing.

  

![bruno image](./docs/images/bruno_auth_api.png)