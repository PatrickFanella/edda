# Automatic frontend deployment

The native Gitea CI workflow retains the backend `task test` job, then builds
and validates the frontend Docker image. Pull requests build without production
credentials. Pushes to `main` and manual dispatches on `main` deploy the exact
verified revision to Dozor's existing `edda-web` container, Compose service `web`.
Other branches do not activate production.

The frontend uses an operator-reviewed service-only Compose baseline derived from
the running container. It retains the existing `projects` network, private LAN
binding on port 3037, restart policy, logging and resource limits. The baseline
lives outside the API's private project directory and reads no API environment
file. The API and database keep their existing release procedure; `make deploy`
backs up the database and runs migrations before a backend cutover. This workflow
automates frontend delivery and does not replace that backend procedure.

CI uses an app-specific `HOMELAB_DEPLOY_KEY` and pinned `HOMELAB_KNOWN_HOSTS`.
The private fleet-deploy receiver accepts only Edda's frontend and the current
main revision. It refuses runtime drift, verifies the activated image, runtime
settings and public HTML, and restores the previous image on failed readiness.
Retries reuse the first accepted image for that revision instead of rebuilding
it. Exact image IDs, diagnostics and rollback commands stay in private host
receipts. The homelab operator owns the key; rotate it by 2027-02-03.

Register and test the host profile and restricted key before merging. Verify
`https://edda.subcult.tv/` and its linked assets against the deployed image after
rollout. The existing Caddy API route and authentication settings stay under
their own runtime configuration.
