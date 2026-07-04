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
