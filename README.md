# devcli

`devcli` is a Go command-line application with OIDC login, logout, whoami, and doctor commands.

## Build

```bash
go build -o devcli ./cmd/devcli
```

Local builds use the default development version:

```bash
./devcli --version
# devcli v0.0.0-dev
```

To test a next-version development build locally, inject a `{version}-dev` value:

```bash
VERSION=v0.2.0-dev
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
go build \
  -ldflags "-X github.com/nkootstra/cli/internal/version.Version=${VERSION} -X github.com/nkootstra/cli/internal/version.Commit=${COMMIT} -X github.com/nkootstra/cli/internal/version.Date=${DATE}" \
  -o devcli \
  ./cmd/devcli
./devcli --version
```

## Commands

```bash
devcli --version # Show the devcli version
devcli -v        # Show the devcli version
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

## CI Secrets and OAuth Configuration

The default CI workflow uses `oauth2-mock-server` and intentionally hardcodes only non-secret, test-only values:

```yaml
DEVCLI_OIDC_ISSUER_URL: "http://localhost:8080"
DEVCLI_OIDC_CLIENT_ID: "devcli"
DEVCLI_OIDC_SCOPES: "openid profile email"
DEVCLI_OIDC_REDIRECT_URL: "http://127.0.0.1:8765/callback"
DEVCLI_OIDC_CALLBACK_LISTEN_ADDR: "127.0.0.1:8765"
```

Do not commit real OAuth credentials, tokens, auth state files, or provider-specific secrets.
For CI that targets a real OIDC provider, store values in GitHub Actions repository or environment secrets/variables and map them into the job environment:

```yaml
env:
  DEVCLI_OIDC_ISSUER_URL: ${{ vars.DEVCLI_OIDC_ISSUER_URL }}
  DEVCLI_OIDC_CLIENT_ID: ${{ secrets.DEVCLI_OIDC_CLIENT_ID }}
  DEVCLI_OIDC_SCOPES: ${{ vars.DEVCLI_OIDC_SCOPES }}
  DEVCLI_OIDC_REDIRECT_URL: ${{ vars.DEVCLI_OIDC_REDIRECT_URL }}
  DEVCLI_OIDC_CALLBACK_LISTEN_ADDR: ${{ vars.DEVCLI_OIDC_CALLBACK_LISTEN_ADDR }}
```

`devcli` uses an authorization-code flow with PKCE and currently requires a client ID, not a client secret.
If a future provider integration needs a client secret, add a dedicated environment variable and keep it in GitHub Actions secrets rather than source code.

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

## Releases

The release workflow builds Linux, macOS, and Windows archives when a tag matching `v*` is pushed.
Version metadata is injected at build time and is visible through `devcli --version` and `devcli -v`.

Release a new version:

1. Make sure `main` contains the commit you want to release.
2. Pick a SemVer tag with a leading `v`, for example `v0.1.0`.
3. Run the local checks:

```bash
go test ./...
go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/ci.yml .github/workflows/release.yml
```

4. Create and push the tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

5. Wait for the GitHub `Release` workflow to finish.

The workflow creates a GitHub release with `.tar.gz` artifacts and `.sha256` checksum files.
It uses the tag as the released version, the build commit as the commit metadata, and the workflow build time as the date metadata.

You can also run the `Release` workflow manually from GitHub Actions with a version input such as `v0.1.0`.
