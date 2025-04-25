# Log Highlighter (`loghl`)

`loghl` is a command-line tool that reads input from `stdin`, highlights lines containing the word "error" (case-insensitive), and outputs the highlighted text to `stdout`.

## Features

- Highlights the word "error" in bright red.
- Processes input line-by-line for efficient memory usage.
- Works seamlessly with Unix pipelines.

## Usage

```bash
cat your-log-file.log | ./loghl
```

## Example

Input:

```log
INFO: Application started
ERROR: Unable to connect to database
INFO: Retrying connection
```

Output:

```log
INFO: Application started
[91mERROR: Unable to connect to database[0m
INFO: Retrying connection
```
