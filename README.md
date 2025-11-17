# CurlMan - Postman CLI Alternative

A powerful, terminal-based alternative to Postman built with Go. CurlMan allows you to import OpenAPI specifications, manage HTTP requests, execute them, and export them as curl commands - all from your terminal.

## Features

### Core Functionality

- **OpenAPI Import**: Import OpenAPI 3.0+ YAML specifications as collections
  - Automatically extracts server URLs, paths, operations, and parameters
  - Converts server variables and parameter defaults to collection variables
  - Generates properly formatted requests with headers and bodies

- **Request Management**: Full CRUD operations for HTTP requests
  - Create, edit, clone, and delete HTTP requests
  - Support for all HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
  - Clone requests with all properties preserved (creates a copy with "_clone" suffix)

- **Full Request Customization**:
  - Modify HTTP methods
  - Edit base URLs and paths separately
  - Add/modify headers (key-value pairs)
  - Add/modify query parameters (key-value pairs)
  - Edit request bodies
  - Add descriptions

### Variable System

- **Collection Variables**: Define variables specific to each collection
- **Global Variables**: Create global variables available across all collections
  - Stored in `~/.curlman/global.json`
  - Accessible from any collection

- **Variable Injection**: Use `{{variable_name}}` syntax in:
  - URLs and paths
  - Headers
  - Query parameters
  - Request bodies

- **Variable Precedence** (lowest to highest priority):
  1. Global Variables (lowest priority)
  2. Collection Variables
  3. Global Environment Variables
  4. Collection Environment Variables (highest priority)

### Environment Management

- **Dual-Environment System**:
  - **Global Environments**: Shareable across all collections
    - Stored in `~/.curlman/environments/` directory
    - Can be saved, loaded, and reused
    - Perfect for shared configurations (dev, staging, prod)
  - **Collection Environments**: Embedded within each collection
    - Stored directly in the collection JSON file
    - Ideal for project-specific configurations
    - Persistent with the collection

- **Environment Operations**:
  - Create, activate, edit, and delete environments
  - Toggle between global and collection environments with `t` key
  - Manage environment-specific variables
  - Clone and rename environments
  - Visual indicators for active environment

### Request Execution

- **HTTP Client**: Execute requests with full variable substitution
  - 30-second timeout
  - Detailed response capture (status, headers, body, duration)
  - Comprehensive error handling

- **Response Management**:
  - View formatted responses with status and headers
  - Save response body to file (`s` key in response view)
  - Request timing/duration tracking

### Export & Persistence

- **Curl Export**: Generate curl commands from requests
  - Automatic variable substitution
  - Proper escaping and formatting
  - Complete header and body inclusion

- **Collection Persistence**:
  - Save and load collections as JSON files
  - Auto-storage in `~/.curlman/` directory
  - Beautiful JSON formatting with indentation

- **Automatic Storage Management**:
  - Collections: `~/.curlman/*.json`
  - Global config: `~/.curlman/global.json`
  - Global environments: `~/.curlman/environments/*.json`
  - Directory auto-created on first use

### User Interface

- **Interactive TUI**: Beautiful terminal UI built with Bubble Tea
  - Vim-like keybindings (j/k) plus arrow keys
  - Color-coded interface (magenta selection, green success, red errors)
  - Context-aware help system
  - Clear visual feedback for all operations

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
- `v` - Manage collection variables
- `g` - Manage global variables
- `e` - Manage environments (toggle between global/collection with `t`)
- `s` - Save collection to JSON (saved to `~/.curlman/`)
- `l` - Load collection from JSON (loaded from `~/.curlman/`)
- `?` - Show help
- `q` - Quit

### Storage Directory

CurlMan automatically creates and uses `~/.curlman/` as the storage directory for all data:

```
~/.curlman/
├── *.json                    # Collection files
├── global.json               # Global variables
└── environments/             # Global environment files
    ├── development.json
    ├── staging.json
    └── production.json
```

When you save or load a collection using just a filename (e.g., `my-collection.json`), it will be saved to or loaded from this directory. You can still use absolute paths if you need to save/load from a different location.

**Storage Locations**:
- **Collections**: Saved as individual JSON files in `~/.curlman/`
- **Global Variables**: Stored in `~/.curlman/global.json`
- **Global Environments**: Stored as separate JSON files in `~/.curlman/environments/`
- **Collection Environments**: Embedded within each collection JSON file (portable with the collection)

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

- `↑/↓` - Navigate fields (Name, Method, URL, Path, Body)
- `enter` - Edit selected field
- `esc` - Back to request detail

### Variables View (Collection or Global)

- `↑/↓` or `j/k` - Navigate variables
- `enter` - Edit variable value
- `n` - Create new variable
- `d` - Delete variable
- `esc` - Back to main view

### Environment Management View

- `↑/↓` or `j/k` - Navigate environments
- `enter` - Select/view environment or create new
- `t` - Toggle between global and collection environments
- `a` - Activate selected environment
- `v` - Manage environment variables
- `e` - Edit environment name
- `s` - Save environment (global environments only)
- `d` - Delete environment
- `c` - Clone environment
- `esc` - Back to main view

### Response View

- `s` - Save response body to file
- `esc` - Back to request detail

## Variables and Environments

### Variables

Define variables in the Variables view and use them in your requests with the `{{variable_name}}` syntax. Variables are automatically injected when executing requests or exporting to curl.

CurlMan supports multiple types of variables with a clear precedence order:

1. **Global Variables** (`g` in main view): Available across all collections
2. **Collection Variables** (`v` in main view): Specific to the current collection
3. **Global Environment Variables**: Set within global environments
4. **Collection Environment Variables**: Set within collection-specific environments (highest priority)

Example:
```
Variable: baseUrl = https://api.example.com
Variable: apiKey = your-api-key-here

Request URL: {{baseUrl}}/users
Request Header: Authorization: Bearer {{apiKey}}
```

### Environments

Environments allow you to maintain different sets of variables for different contexts (e.g., development, staging, production).

**Global Environments**:
- Shared across all collections
- Stored in `~/.curlman/environments/`
- Useful for organization-wide configurations
- Can be saved and loaded independently

**Collection Environments**:
- Embedded within each collection file
- Travel with the collection when shared
- Ideal for project-specific settings

**Toggle between types**: Press `t` in the environment management view to switch between global and collection environments.

**Activating Environments**: Select an environment and press `a` to activate it. Active environment variables will override collection and global variables.

## Example Workflows

### Basic Workflow: Import and Execute

1. **Import an OpenAPI file**:
   - Press `i` in the main view
   - Enter the path to your OpenAPI YAML file (e.g., `example.yaml`)
   - CurlMan automatically extracts all endpoints and creates requests

2. **View and select a request**:
   - Press `r` to view requests
   - Navigate with arrow keys or `j/k`
   - Press `enter` to view details

3. **Customize the request**:
   - Press `e` to edit basic fields (name, method, URL, path, body)
   - Press `h` to add/modify headers
   - Press `p` to add/modify query parameters

4. **Execute the request**:
   - In request detail view, press `enter` to execute
   - View the response with status, headers, and body
   - Press `s` to save the response body to a file (optional)

5. **Export as curl**:
   - In request detail view, press `x`
   - Copy the generated curl command with all variables substituted

6. **Save your collection**:
   - Go to main view (press `esc` until you're back)
   - Press `s` to save
   - Enter filename (e.g., `my-api.json`)

### Advanced Workflow: Multi-Environment Setup

1. **Set up global variables**:
   - Press `g` in the main view
   - Press `n` to create a new variable
   - Add common variables (e.g., `apiVersion = v1`)

2. **Create environments for different stages**:
   - Press `e` to access environment management
   - Press `enter` to create a new global environment
   - Name it "Development"
   - Press `v` to manage variables, then add:
     - `baseUrl = https://dev-api.example.com`
     - `apiKey = dev-key-12345`
   - Press `esc` to go back

3. **Create production environment**:
   - Press `enter` again to create another environment
   - Name it "Production"
   - Press `v` and add:
     - `baseUrl = https://api.example.com`
     - `apiKey = prod-key-67890`

4. **Activate and switch environments**:
   - Navigate to "Development" environment
   - Press `a` to activate it
   - Your requests now use dev variables
   - To switch to production: navigate to "Production" and press `a`

5. **Create collection-specific environments**:
   - Press `t` to toggle to collection environments
   - Create environments that will be saved with the collection
   - Perfect for project-specific overrides

6. **Use variables in requests**:
   - Edit a request (press `r`, select request, press `e`)
   - Set URL to `{{baseUrl}}`
   - Add header: `Authorization: Bearer {{apiKey}}`
   - The active environment variables will be injected automatically

### Workflow: Sharing Collections with Teams

1. **Save collection with embedded environments**:
   - Create collection environments (`e`, then `t` to toggle to collection)
   - Add all project-specific variables
   - Save the collection (`s`)

2. **Share the JSON file**:
   - The collection file includes all requests and collection environments
   - Team members can load it (`l`) and immediately have the same setup

3. **Use global environments for personal settings**:
   - Each team member can create their own global environments
   - Store personal API keys or local server URLs
   - These won't be shared with the collection file

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
├── storage/         # Storage directory management
├── main.go          # Application entry point
├── example.yaml     # Sample OpenAPI specification
└── README.md        # This file
```

## Data Models

Understanding the data structures used by CurlMan:

### Request Model
Each request contains:
```json
{
  "id": "unique-uuid",
  "name": "Get User Profile",
  "method": "GET",
  "url": "{{baseUrl}}",
  "path": "/api/users/{{userId}}",
  "headers": {
    "Authorization": "Bearer {{apiKey}}",
    "Content-Type": "application/json"
  },
  "queryParams": {
    "include": "profile,settings"
  },
  "body": "",
  "description": "Fetches user profile information"
}
```

### Collection Model
Collections organize requests and variables:
```json
{
  "name": "My API Collection",
  "requests": [...],
  "variables": {
    "baseUrl": "https://api.example.com",
    "userId": "12345"
  },
  "activeEnvironment": "development",
  "environmentVars": {},
  "environments": [
    {
      "name": "development",
      "variables": {
        "baseUrl": "https://dev-api.example.com",
        "apiKey": "dev-key"
      }
    },
    {
      "name": "production",
      "variables": {
        "baseUrl": "https://api.example.com",
        "apiKey": "prod-key"
      }
    }
  ],
  "activeCollectionEnv": "development",
  "collectionEnvVars": {}
}
```

### Global Variables File (`~/.curlman/global.json`)
```json
{
  "variables": {
    "apiVersion": "v1",
    "timeout": "30s"
  }
}
```

### Environment File (`~/.curlman/environments/development.json`)
```json
{
  "name": "development",
  "variables": {
    "baseUrl": "https://dev-api.example.com",
    "apiKey": "dev-api-key-12345",
    "debugMode": "true"
  }
}
```

## Tips and Best Practices

### Organizing Your Workflow

1. **Use Global Variables for Constants**: Store API versions, common timeouts, or organization-wide values in global variables (`g`)
2. **Use Collection Variables for Project Specifics**: Store base URLs and project-specific values in collection variables (`v`)
3. **Use Environments for Deployment Stages**: Create separate environments for dev, staging, and production

### Environment Strategy

**Global Environments** - Use when:
- Working across multiple collections that share the same infrastructure
- You want to switch all projects between dev/staging/prod at once
- Sharing environment configurations across a team (via separate files)

**Collection Environments** - Use when:
- Each collection needs different environment values
- You want to share a complete collection setup with someone
- Working on a project that has unique deployment configurations

### Variable Naming Conventions

Recommended naming patterns:
- `baseUrl` - Base API URL
- `apiKey` - Authentication keys
- `apiVersion` - API version number
- `userId`, `orderId` - Entity identifiers
- `timeout` - Request timeout values

### Security Best Practices

1. **Never commit API keys**: Use environment variables for sensitive data
2. **Use collection environments for examples**: Share collections with placeholder values in collection environments
3. **Keep personal keys in global environments**: Store your actual API keys in global environments (not saved with collections)
4. **Separate files for different security levels**: Keep development keys in version control, production keys separate

### Performance Tips

1. **Use request cloning** (`c` key) to quickly create similar requests
2. **Save responses** (`s` in response view) for offline analysis
3. **Export to curl** (`x` key) for quick command-line execution
4. **Organize requests by feature** within collections for easy navigation

### Keyboard Efficiency

- Use `j/k` for Vim-style navigation (faster than arrow keys)
- Learn the main shortcuts: `r` (requests), `v` (variables), `e` (environments), `g` (global)
- Use `esc` liberally to navigate back through views
- Press `?` anytime to see contextual help

## Troubleshooting

### Common Issues

**Problem**: Variables not being replaced in requests
- **Solution**: Check variable precedence - ensure the variable is defined and the environment (if any) is activated with `a`

**Problem**: Can't find saved collections
- **Solution**: Collections are saved to `~/.curlman/` by default. Check this directory or use absolute paths

**Problem**: Request fails with timeout
- **Solution**: Default timeout is 30 seconds. Check your network connection and API endpoint availability

**Problem**: OpenAPI import creates malformed requests
- **Solution**: Ensure your OpenAPI file is valid YAML and follows OpenAPI 3.0+ specification

**Problem**: Environment changes not taking effect
- **Solution**: Make sure you've activated the environment with `a` after selecting it

### File Locations

If you need to manually inspect or edit files:
- Collections: `~/.curlman/*.json`
- Global variables: `~/.curlman/global.json`
- Global environments: `~/.curlman/environments/*.json`

All files are standard JSON and can be edited with any text editor.

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
