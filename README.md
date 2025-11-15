# CurlMan - Postman CLI Alternative

A powerful, terminal-based alternative to Postman built with Go. CurlMan allows you to import OpenAPI specifications, manage HTTP requests, execute them, and export them as curl commands - all from your terminal.

## Features

- **OpenAPI Import**: Import OpenAPI YAML files as collections
- **Request Management**: Create, edit, clone, and delete HTTP requests
- **Full Request Customization**:
  - Modify HTTP methods (GET, POST, PUT, DELETE, etc.)
  - Edit URLs and paths
  - Add/modify headers
  - Add/modify query parameters
  - Edit request bodies
- **Variable Injection**: Define variables and inject them into requests using `{{variable_name}}` syntax
- **Request Execution**: Execute HTTP requests directly from the terminal
- **Curl Export**: Export any request as a curl command
- **Collection Persistence**: Save and load collections as JSON files
- **Interactive TUI**: Beautiful terminal UI built with Bubble Tea

## Installation

```bash
go build -o curlman .
```

## Usage

Run the application:

```bash
./curlman
```

### Main View Commands

- `i` - Import OpenAPI YAML file
- `r` - View requests
- `v` - Manage variables
- `s` - Save collection to JSON
- `l` - Load collection from JSON
- `?` - Show help
- `q` - Quit

### Request List View

- `↑/↓` or `j/k` - Navigate requests
- `enter` - View request details
- `n` - Create new request
- `d` - Delete selected request
- `esc` - Back to main

### Request Detail View

- `enter` - Execute request
- `e` - Edit request
- `h` - Manage headers
- `p` - Manage query parameters
- `c` - Clone request
- `x` - Export as curl command
- `esc` - Back to request list

### Request Edit View

- `↑/↓` - Navigate fields
- `enter` - Edit selected field
- `esc` - Back to request detail

### Variables

Define variables in the Variables view and use them in your requests with the `{{variable_name}}` syntax. Variables are automatically injected when executing requests or exporting to curl.

Example:
```
Variable: baseUrl = https://api.example.com
Variable: apiKey = your-api-key-here

Request URL: {{baseUrl}}/users
Request Header: Authorization: Bearer {{apiKey}}
```

## Example Workflow

1. **Import an OpenAPI file**:
   - Press `i` in the main view
   - Enter the path to your OpenAPI YAML file (e.g., `example.yaml`)

2. **View and select a request**:
   - Press `r` to view requests
   - Navigate with arrow keys
   - Press `enter` to view details

3. **Customize the request**:
   - Press `e` to edit basic fields
   - Press `h` to add/modify headers
   - Press `p` to add/modify query parameters

4. **Set up variables**:
   - Go back to main view (press `esc` twice)
   - Press `v` to manage variables
   - Press `enter` to add a variable

5. **Execute the request**:
   - Navigate to a request detail view
   - Press `enter` to execute
   - View the response with status, headers, and body

6. **Export as curl**:
   - In request detail view, press `x`
   - Copy the generated curl command

7. **Save your collection**:
   - Go to main view
   - Press `s` to save
   - Enter filename (e.g., `my-collection.json`)

## Libraries Used

- [kin-openapi](https://github.com/getkin/kin-openapi) - OpenAPI specification parser
- [bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal output
- [bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## Project Structure

```
curlman/
├── models/          # Data structures (Collection, Request)
├── openapi/         # OpenAPI import/export functionality
├── ui/              # Bubble Tea TUI implementation
├── executor/        # HTTP request execution
├── exporter/        # Curl command generation
├── main.go          # Application entry point
├── example.yaml     # Sample OpenAPI specification
└── README.md        # This file
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
