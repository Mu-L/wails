# Wails v3 documentation

The M-Press migration lives in `docs/mpress/`, with English as the default
language at `/` and existing translations under their language codes.
[Correction PRs are welcome](mpress/CONTRIBUTING.md), including typos, links,
examples, and translations. No issue or failing code test is required for a
documentation-only correction.

## Preview and validate

Install [M-Press v1.0.1](https://github.com/leaanthony/mpress/releases/tag/v1.0.1)
from a release archive, or use Go:

```sh
go install github.com/leaanthony/mpress/cmd/mpress@v1.0.1
```

From the repository root:

```sh
mpress dev
mpress build --strict --no-purge-css
mpress check
python3 docs/mpress/scripts/check_site.py docs/mpress/site
```

The root `mpress.yaml` points to `docs/mpress/content`, static assets, and CSS.
This entry point lets the contribution command work from a cloned repository.
`docs/mpress/mpress.yaml` also supports a standalone build from that directory;
its local contribution wizard is disabled because its paths are project-relative.
Keep both configurations aligned when changing site settings.

English source files live in `docs/mpress/content/`; translated files live under
folders such as `fr/`, `id/`, and `zh-cn/`. Output goes to `docs/mpress/site/`.
The build retains CSS because the imported home-page animation creates classes
at runtime. Contributors do not need Cloudflare credentials, Node.js, private
services, or a translation provider.

## Cloudflare build

Run `bash docs/mpress/scripts/build.sh` on Linux AMD64 to execute the exact CI
build: download M-Press v1.0.1, verify its pinned SHA-256 digest, build, and check.
The output is a static directory suitable for Cloudflare Pages Direct Upload.

The `Wails v3 documentation` workflow validates PRs without deployment secrets.
An approved `docs-cloudflare` environment with `CLOUDFLARE_PAGES_API_TOKEN` enables
deployment from trusted `master` pushes or a manual preview run once the repository
variable `WAILS_DOCS_DEPLOY_ENABLED` is set to `true`. The token must
have Pages Edit access only to the intended account; never copy a personal
Wrangler OAuth token into CI. Production cutover is separate from this migration.

## Starlight rollback source

`docs/src/content/docs/`, `astro.config.mjs`, and the npm lockfile remain intact.
The migration was taken from Wails commit
`4146200f1` and supersedes the static artifact in PR #5905.
Before production cutover, reconcile any later Starlight/changelog changes and
switch documentation-producing automation to the migrated source.

To reproduce the old site, install D2, then run `npm ci && npm run build` from
`docs/`. Its output is `docs/dist/` and is not the M-Press deployment output.
