# devcli

`devcli` is a Go command-line application with OIDC login, logout, whoami, and doctor commands.

## Build

```bash
go build -o devcli ./cmd/devcli
```

## Commands

```bash
devcli login   # Authenticate with an OIDC provider using auth code + PKCE
devcli whoami  # Show the current logged-in identity
devcli logout  # Remove the saved local session
devcli doctor  # Check configuration, auth state, and OIDC discovery
```

Each command supports `--help` with its flags and description.

## OIDC Configuration

Settings can be supplied by flags or environment variables. Flags take precedence over environment variables.

| Setting | Environment variable | Flag |
|---|---|---|
| OIDC issuer URL | `DEVCLI_OIDC_ISSUER_URL` | `--oidc-issuer-url` |
| OIDC client ID | `DEVCLI_OIDC_CLIENT_ID` | `--oidc-client-id` |
| OIDC scopes | `DEVCLI_OIDC_SCOPES` | `--oidc-scopes` |
| Redirect URL | `DEVCLI_OIDC_REDIRECT_URL` | `--oidc-redirect-url` |
| Callback listen address | `DEVCLI_OIDC_CALLBACK_LISTEN_ADDR` | `--oidc-callback-listen-addr` |
| Auth state path | `DEVCLI_AUTH_STATE_PATH` | `--auth-state-path` |
| Mock issuer URL | `DEVCLI_MOCK_OIDC_ISSUER_URL` | `--mock-oidc-issuer-url` |

Example local configuration:

```bash
export DEVCLI_OIDC_ISSUER_URL=http://localhost:8080
export DEVCLI_OIDC_CLIENT_ID=devcli
export DEVCLI_OIDC_SCOPES="openid profile email"
export DEVCLI_OIDC_REDIRECT_URL=http://127.0.0.1:8765/callback
export DEVCLI_OIDC_CALLBACK_LISTEN_ADDR=127.0.0.1:8765
```

`devcli login` stores token-bearing state under the user config directory by default.
Use `DEVCLI_AUTH_STATE_PATH` or `--auth-state-path` to isolate state for tests.
Do not commit auth state files.

## Local OIDC Mock Server

The repository includes Docker Compose support for `oauth2-mock-server`.
It is for local development and CI only.

Start the mock OIDC provider:

```bash
docker compose up --build -d oauth2-mock-server
```

The mock server advertises `http://localhost:8080` as its issuer URL, so use `localhost` rather than `127.0.0.1` for `DEVCLI_OIDC_ISSUER_URL`.

## Developer OAuth Test

Use this flow to test the CLI manually against the local OAuth server.

1. Start the mock server:

```bash
docker compose up --build -d oauth2-mock-server
```

2. Export local OIDC settings:

```bash
export DEVCLI_OIDC_ISSUER_URL=http://localhost:8080
export DEVCLI_OIDC_CLIENT_ID=devcli
export DEVCLI_OIDC_SCOPES="openid profile email"
export DEVCLI_OIDC_REDIRECT_URL=http://127.0.0.1:8765/callback
export DEVCLI_OIDC_CALLBACK_LISTEN_ADDR=127.0.0.1:8765
```

3. Build the CLI:

```bash
go build -o devcli ./cmd/devcli
```

4. Check configuration and login:

```bash
./devcli doctor
./devcli login
```

`login` prints an authorization URL.
Open it in your browser; the mock server redirects back to the local callback listener and `devcli` saves the verified session.

5. Verify and clear the session:

```bash
./devcli whoami
./devcli logout
```

6. Stop the mock server:

```bash
docker compose down
```

## Tests

Unit tests do not require Docker:

```bash
go test ./...
```

Run the Docker-backed OIDC integration test with:

```bash
docker compose up --build -d oauth2-mock-server
DEVCLI_RUN_INTEGRATION=1 \
DEVCLI_OIDC_ISSUER_URL=http://localhost:8080 \
DEVCLI_OIDC_CLIENT_ID=devcli \
DEVCLI_OIDC_SCOPES="openid profile email" \
DEVCLI_OIDC_REDIRECT_URL=http://127.0.0.1:8765/callback \
DEVCLI_OIDC_CALLBACK_LISTEN_ADDR=127.0.0.1:8765 \
go test ./test/integration -v
```

CI runs unit tests first, starts the mock provider, waits for OIDC discovery, and then runs the integration test with environment variables.
