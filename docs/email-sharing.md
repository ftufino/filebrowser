# Email sharing

Administrators can configure outgoing email from **Settings → Global Settings → Outgoing email (SMTP)**:

- **SMTP host**: server hostname (without a port).
- **SMTP port**: defaults to `587`.
- **SMTP username and password**: optional, but configure both or neither. A blank password keeps the currently saved password.
- **Sender email address**: the address shown as the sender.

File Browser requires STARTTLS and verifies the SMTP server certificate. Configure a STARTTLS endpoint (commonly port `587`); implicit-TLS endpoints such as port `465` are not supported. Leave the host empty to disable email sharing, which also clears the stored SMTP credentials. SMTP passwords are never returned by the settings API.

Users with both **Share** and **Download** permissions can select a file, open **Share**, enter a recipient address, and choose **Send email**. The selected file is attached directly; folders cannot be sent, and attachments are limited to 25 MiB. This sends the file contents, not a share link.

SMTP credentials are stored with the application settings. Restrict access to configuration exports and backups accordingly.
