generic portfolio/blog API

## Local development

Both flags are mandatory — the API exits without them:

```bash
go run . --config config.json --log /tmp/argsea-api.log
```

The server listens on port **8181**.

`config.json` is gitignored (it holds live credentials) — copy
`config.example.json` and fill it in. Mongo is not local: it is reached over an
SSH tunnel to the VPS, so open the tunnel first:

```bash
ssh -N -L 27017:localhost:27017 argsea
```

### Trying it

```bash
# public read: published projects only (what the Astro build consumes)
curl 'http://127.0.0.1:8181/1/project?published=true'

# public read: the site-copy singleton ("signal flags")
curl 'http://127.0.0.1:8181/1/copy'

# authed write: create a project. Writes accept the auth-token session cookie
# or an Authorization: Bearer JWT — use Bearer locally (the cookie is scoped to
# argsea.com). Without either the API answers 401.
curl -X POST 'http://127.0.0.1:8181/1/project' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <jwt>' \
  -d '{"title":"Postcard","category":"backend","shortDesc":"hello","status":"draft"}'
```

## The lantern (deploy-on-hoist)

Admin-only deploy endpoint: it runs the site build, moves `dist/` into a
timestamped generation under `releases_dir`, atomically re-points the
`live_link` symlink (what nginx serves), prunes old generations, and records
the hoist in the activity log. The previous generation stays on disk for
instant rollback.

**The `lantern` config section is the feature flag** — when it is absent the
routes are not mounted at all. See `config.example.json` for the shape:
`site_dir` (build working directory), `build_cmd` (an argv array, never a
shell string), `dist_dir` (relative to `site_dir`), `releases_dir`,
`live_link`, `keep` (generations to retain, default 2), `timeout_seconds`
(default 600), and `env` — an array of `KEY=VALUE` strings merged over the
process env for the build. It is an array, not a JSON object, because the
config loader lowercases object keys (`ARGSEA_API_URL` would silently become
`argsea_api_url`).

Both routes require a JWT whose `role` claim is `admin` — 401 without a valid
token, 403 with a plain-user one. Login mints the role stored on the user
document; roles are never accepted from request bodies, so admin is granted
only by a direct DB update.

```bash
# start a hoist — 202 with the fresh status, 409 if one is already running
curl -X POST 'http://127.0.0.1:8181/1/lantern/hoist/' -H 'Authorization: Bearer <admin jwt>'

# poll it — state: idle|building|swapping|succeeded|failed, plus startedAt,
# finishedAt, lastHoistedAt, and a bounded tail of build output
curl 'http://127.0.0.1:8181/1/lantern/' -H 'Authorization: Bearer <admin jwt>'
```
