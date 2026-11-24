# File Browser

Web UI served on the configured port (default `8080`). Default login: `admin` / `admin` (change it immediately).
Files are served from `/share`; `/media` is also mounted.
Database and settings are stored in the add-on `/data` folder.

## Options

- `port`: port the service listens on (the add-on uses host networking).
- `base_url`: path prefix, e.g. `/files`, for use behind a reverse proxy that forwards that prefix unchanged. Leave empty to serve at the root.

Example nginx: `location /files/ { proxy_pass http://192.168.1.200:8080; }` with `base_url: /files`.
