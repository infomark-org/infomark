# InfoMark console

The backend includes a console for common tasks such as creating and restoring
a backup, manually confirming user accounts, and setting permissions.

> Currently, the console can only be accessed directly on the server.

For a full, always up-to-date list of commands and their flags, run the binary
with `--help`, for example:

```bash
./infomark console --help
./infomark console user --help
./infomark console course enroll --help
```

The most common console commands are covered in context throughout the other
guides:

- The [Overview](overview.md) walks through `configuration`, `user`, and
  `admin` commands when setting up an instance and its first user.
- The [Administrator's Guide](administrator.md) covers `admin`, `course`,
  `group`, and `database` (backup and restore) commands.
