# sloghandler

`sloghandler` is a Go package that provides a custom log handler for the [slog](https://pkg.go.dev/log/slog) logging library. It enables formatted log output similar to the default log package but with support for structured attributes, log levels, thread-safe writes, and optional grouping.

---

## Features

- **Formatted Logs**: Logs are formatted with timestamps, log levels, and messages in a readable format.
- **Structured Logging**: Supports key-value pairs for structured log attributes.
- **Thread-Safe Writes**: Ensures safe concurrent logging in multi-threaded applications.
- **Customizable Attributes**: Add global attributes or group logs with a prefix.
- **Level Filtering**: Handle logs only above a specified log level.
