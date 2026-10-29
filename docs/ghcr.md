# GitHub Container Registry

The release workflow publishes multi-architecture images to `ghcr.io/ftufino/filebrowser` when a `v*` release tag is pushed. It publishes the version tag, the major-version tag, and `latest` for the standard image; the s6 image is available with the `-s6` suffix and as `s6`.

Examples:

```sh
docker pull ghcr.io/ftufino/filebrowser:latest
docker pull ghcr.io/ftufino/filebrowser:v3.0.0
docker pull ghcr.io/ftufino/filebrowser:v3.0.0-s6
```

The workflow uses the repository's `GITHUB_TOKEN` with `packages: write` permission. After its first successful publication, set the package visibility to **Public** in the package settings if it should be pullable without authentication.
