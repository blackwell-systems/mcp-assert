# Rust Filesystem Server Examples

These assertions test [rust-mcp-stack/rust-mcp-filesystem](https://github.com/rust-mcp-stack/rust-mcp-filesystem) (145 stars), a Rust reimplementation of the official TypeScript filesystem server. Built on the active `rust-mcp-sdk`.

## Setup

```bash
cargo install rust-mcp-filesystem
```

Then run:

```bash
mcp-assert run --suite examples/rust-filesystem --fixture examples/rust-filesystem/fixtures
```

## Coverage

23 assertions covering 22/24 tools (92%):

| Category | Tools | Assertions |
|----------|-------|------------|
| Read | `read_text_file`, `read_multiple_text_files`, `head_file`, `tail_file`, `read_file_lines` | 5 |
| List | `list_directory`, `list_directory_with_sizes`, `directory_tree`, `list_allowed_directories` | 4 |
| Search | `search_files`, `search_files_content`, `find_empty_directories`, `find_duplicate_files` | 4 |
| Info | `get_file_info`, `calculate_directory_size` | 2 |
| Write | `write_file`, `edit_file`, `create_directory`, `move_file` | 4 |
| Zip | `zip_files`, `zip_directory`, `unzip_file` | 3 |
| Security | Path traversal rejection (negative test) | 1 |

Not covered: `read_media_file`, `read_multiple_media_files` (require binary fixtures).

## Notes

Clean scan: no bugs found. All 23 assertions pass. The server correctly rejects path traversal attempts outside the allowed directory.

Write operations use `--allow-write` flag and fixture isolation to prevent contaminating the fixture directory between test runs.
