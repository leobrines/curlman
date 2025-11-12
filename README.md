# curlman

An interactive CLI tool for API testing as an alternative to Postman. Built with Go and designed for developers who prefer working in the terminal.

## Features

- **Interactive TUI**: Beautiful terminal interface built with [Bubbletea](https://github.com/charmbracelet/bubbletea)
- **Environments**: Manage multiple environments with key-value variables
- **Collections**: Organize requests in collections (supports OpenAPI files)
- **Dual Request Types**:
  - **Spec Requests**: Generated on-the-fly from OpenAPI files (ephemeral, read-only)
  - **Managed Requests**: User-saved requests with full control (editable, persistent)
- **Templating**: Use Go templates with environment variables in your requests
- **Vim Integration**: Edit requests and collections using Vim (or your preferred editor)
- **Request Execution**: Execute curl commands with real-time output
- **OpenAPI Support**: Import OpenAPI specifications as in-memory collections

## Installation

### Quick Start

```bash
# Clone the repository
git clone https://github.com/leit0/curlman.git
cd curlman

# Install
make install
```

### Manual Installation

```bash
# Build from source
go build -o curlman cmd/curlman/main.go

# Move to PATH
sudo mv curlman /usr/local/bin/
```

### Dependencies

curlman requires Go 1.21 or later. Dependencies are managed via Go modules and will be automatically installed during build.

## Usage

### Initialize

First, initialize curlman in your project directory:

```bash
curlman -i
# or
curlman --init
```

This creates a `.curlman` directory with the following structure:

```
.curlman/
├── config.json          # Application configuration
├── environments/        # Environment files
├── collections/         # Collection files
└── requests/           # Request curl files
```

### Interactive Mode

Launch the interactive TUI (default behavior):

```bash
curlman
```

**Keyboard Shortcuts:**
- `Tab`: Switch between panels (Environments → Collections → Requests)
- `Shift+Tab`: Switch panels backwards
- `↑/↓` or `k/j`: Navigate within a panel
- `Enter`: Select environment/collection or execute request
- `e`: Edit selected item (opens Vim) - managed requests only
- `a`: Add new item
- `d`: Delete selected item - managed requests only
- `s`: Save spec request as managed request
- `c`: Copy curl command to clipboard
- `r`: Refresh OpenAPI spec (reload endpoints)
- `Esc`: Cancel execution or return to list view
- `Ctrl+C` (twice): Exit application

### Wrap Curl Commands

Convert existing curl commands into curlman requests:

```bash
curlman -s 'curl -X GET https://api.example.com/users'
# or
curlman --wrap 'curl -X GET https://api.example.com/users'
```

This will:
1. Validate the curl command
2. Open the interactive UI
3. Prompt you to select a collection
4. Save the request

### Import OpenAPI

Load an OpenAPI file as a collection:

```bash
curlman --openapi api-spec.yaml
```

This will:
1. Parse the OpenAPI specification
2. Create a collection with a reference to the OpenAPI file
3. Launch the interactive UI

**How OpenAPI Collections Work:**
- When you select an OpenAPI collection, **spec requests** are generated on-the-fly in memory
- Spec requests are marked with `[Spec]` prefix and are **read-only**
- The OpenAPI file is **not copied** - it's referenced by its absolute path
- Changes to the OpenAPI file are reflected when you press `r` (refresh)

**Saving Spec Requests:**
- Select any spec request and press `s` to save it as a **managed request**
- Managed requests are editable, persistent, and support full features
- You can save multiple variations of the same endpoint

## Environments

Environments store key-value pairs that can be used as template variables in your requests.

**Example Environment:**
```json
{
  "name": "Production",
  "variables": {
    "BaseURL": "https://api.production.com",
    "ApiKey": "your-api-key-here",
    "Version": "v1"
  }
}
```

**Using in Requests:**
```bash
curl -X GET "{{.BaseURL}}/{{.Version}}/users" \
  -H "Authorization: Bearer {{.ApiKey}}"
```

## OpenAPI Collections

OpenAPI collections provide a dynamic way to work with API specifications.

### Spec Requests vs Managed Requests

**Spec Requests** (marked with `[Spec]` prefix):
- Generated automatically from the OpenAPI file
- Always reflect the current state of the OpenAPI spec
- Read-only (cannot be edited or deleted)
- Ephemeral (exist only in memory, not saved to disk)
- Updated when you press `r` to refresh

**Managed Requests** (no prefix):
- Saved by the user (press `s` on a spec request)
- Fully editable and persistent
- Support request history and response tracking
- Can be customized with different parameter values
- Linked to their source OpenAPI operation (if saved from spec)

### Working with OpenAPI Collections

**1. View Spec Requests:**
   - Select an OpenAPI collection
   - Browse all available endpoints as spec requests
   - Press `Enter` to execute immediately

**2. Save a Spec Request:**
   - Select a spec request
   - Press `s` to save as managed request
   - The managed request becomes editable and persistent

**3. Edit Managed Requests:**
   - Select a managed request (no `[Spec]` prefix)
   - Press `e` to open in your editor
   - Make changes and save

**4. Refresh OpenAPI Spec:**
   - Press `r` while viewing an OpenAPI collection
   - Spec requests update to reflect file changes
   - Managed requests remain unchanged

**5. Operation Validation:**
   - If an OpenAPI operation is removed/changed, managed requests linked to it show `⚠` warning
   - The request still works, but indicates the source endpoint has changed

### Example Workflow

```bash
# 1. Import OpenAPI file
curlman --openapi petstore.yaml

# 2. In the UI:
#    - Select the collection
#    - Browse spec requests (marked with [Spec])
#    - Press Enter on "GET /pets" to test it

# 3. Save commonly used endpoints:
#    - Select "[Spec] List all pets"
#    - Press 's' to save as managed request
#    - Press 'e' to edit and customize parameters

# 4. If petstore.yaml changes:
#    - Press 'r' to refresh spec requests
#    - Saved managed requests remain unchanged
```

## Collections

Collections are groups of related API requests. There are two types:

**1. Standard Collections:**
   - Contain only managed requests
   - Created manually with `a` key
   - All requests are editable and persistent

**2. OpenAPI Collections:**
   - Linked to an OpenAPI specification file
   - Display both spec requests (from OpenAPI) and managed requests (saved)
   - Spec requests regenerated on-the-fly when collection is selected
   - Marked with "(OpenAPI)" in the collection list

**Adding a Standard Collection:**
1. Press `a` in the Collections panel
2. Enter collection name and description
3. Start adding requests with `a` in Requests panel

**Working with Mixed Collections:**
- OpenAPI collections show both request types
- `[Spec]` requests appear first, followed by your saved managed requests
- Both types can be executed with `Enter`
- Only managed requests can be edited (`e`) or deleted (`d`)

## Requests

Each request is a curl command stored in a file within your collection.

**Request File Example** (`.curlman/requests/{collection_id}/{request_id}.curl`):
```bash
curl -X POST "{{.BaseURL}}/api/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {{.ApiKey}}" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }'
```

## Template Engine

curlman uses Go's `text/template` engine for variable substitution.

**Template Syntax:**
- `{{.VariableName}}` - Insert environment variable
- `{{.BaseURL}}/path` - Concatenate with strings

**Example:**
```bash
# Environment variable: BaseURL = "https://api.example.com"
# Template: {{.BaseURL}}/users
# Result: https://api.example.com/users
```

## Directory Structure

```
.curlman/
├── config.json
│   └── Stores selected environment and collection
├── environments/
│   ├── {env-id-1}.json
│   └── {env-id-2}.json
├── collections/
│   ├── {collection-id-1}.json          # Standard collection
│   └── {collection-id-2}.json          # OpenAPI collection (contains OpenAPIPath reference)
└── requests/
    ├── {collection-id-1}/
    │   ├── {request-id-1}.curl         # Managed requests only
    │   └── {request-id-2}.curl         # (Spec requests are NOT stored here)
    └── {collection-id-2}/
        └── {request-id-1}.curl         # Saved from spec request
```

**Important Notes:**
- `requests/` directory contains **only managed requests** (saved by user)
- **Spec requests** from OpenAPI files are generated in-memory and never saved to disk
- OpenAPI collections store a reference path to the external .yaml/.json file
- When you refresh (`r`), spec requests regenerate from the referenced OpenAPI file

## Advanced Usage

### Custom Editor

Set your preferred editor via environment variable:

```bash
export EDITOR=nvim
curlman
```

### Git Integration

Add `.curlman` to your repository to share collections with your team:

```bash
git add .curlman
git commit -m "Add API collections"
git push
```

**Note:** Consider using `.gitignore` for sensitive data:

```
.curlman/environments/*
!.curlman/environments/example.json
```

## Development

### Project Structure

```
curlman/
├── cmd/
│   └── curlman/          # Main application entry
├── internal/
│   ├── config/           # Configuration management
│   ├── storage/          # File system operations
│   ├── models/           # Data structures
│   ├── parser/           # Curl parser
│   ├── executor/         # Request executor
│   ├── openapi/          # OpenAPI parser
│   ├── template/         # Template engine
│   └── ui/               # Bubbletea TUI
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Building

```bash
# Build for current platform
make build

# Install to /usr/local/bin
make install

# Run tests
make test

# Clean build artifacts
make clean
```

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Author

Created by [leit0](https://github.com/leit0)

## Acknowledgments

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [go-openapi](https://github.com/go-openapi) - OpenAPI parsing
