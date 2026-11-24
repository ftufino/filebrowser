# File Browser

Web UI on port 8080 inside the add-on; change the exposed port in the add-on **Network** section. Default login: `admin` / `admin` (change it immediately).
Files are served from `/share`; `/media` is also mounted.
Database and settings are stored in the add-on `/data` folder.

## Options

- `base_url`: path prefix, e.g. `/files`, for use behind a reverse proxy that forwards that prefix unchanged. Leave empty to serve at the root.

Example nginx: `location /files/ { proxy_pass http://192.168.1.200:8080; }` with `base_url: /files`.
