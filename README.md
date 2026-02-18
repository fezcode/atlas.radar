# atlas.radar

**atlas.radar** is a fast, minimalist Git status monitoring tool. It scans a directory for subdirectories containing `.git` repositories and provides a real-time summary of their current status.

## Features

- **Concise UI:** Minimalist, color-coded status reports.
- **Remote Tracking:** Automatically shows `ahead` (↑) and `behind` (↓) counts.
- **Change Counting:** Displays counts for added (+), modified (~), and deleted (-) files.
- **Filtering:** Use `--show` to filter by clean or unclean repositories.
- **Watch Mode:** Use `--watch` to continuously monitor your projects.

## Usage

```bash
# Basic usage (scans the current directory)
atlas.radar

# Scan a specific directory
atlas.radar D:\Workhammer

# Show only unclean repositories
atlas.radar --show unclean

# Continuous monitoring
atlas.radar --watch
```

### Options

| Option | Description | Values | Default |
| :--- | :--- | :--- | :--- |
| `--show` | Filter repository display | `all`, `clean`, `unclean` | `all` |
| `--watch` | Monitor status continuously | `true`, `false` | `false` |
| `--table` | Display results in a table | `true`, `false` | `false` |
| `--pattern` | Regex pattern to match repository names | `string` | `""` |
| `--fetch` | Fetch all updates from remotes | `true`, `false` | `false` |
| `--pull` | Pull all updates from remotes | `true`, `false` | `false` |
| `--push` | Push all local updates to remotes | `true`, `false` | `false` |

## Bulk Operations

You can perform Git operations across all detected repositories:

```bash
# Fetch updates for all projects
atlas.radar --fetch

# Pull updates for all projects
atlas.radar --pull

# Push updates for all projects
atlas.radar --push
```

## Build

Built with [gobake](https://github.com/fezcode/gobake).

```bash
gobake build
```

## License

MIT
