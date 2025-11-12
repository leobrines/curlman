package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/leit0/curlman/internal/config"
	"github.com/leit0/curlman/internal/openapi"
	"github.com/leit0/curlman/internal/parser"
	"github.com/leit0/curlman/internal/storage"
	"github.com/leit0/curlman/internal/ui"
)

var (
	initFlag    bool
	wrapCmd     string
	openapiFile string
)

func init() {
	flag.BoolVar(&initFlag, "i", false, "Initialize .curlman directory in current path")
	flag.BoolVar(&initFlag, "init", false, "Initialize .curlman directory in current path")
	flag.StringVar(&wrapCmd, "s", "", "Wrap a curl command and add it to a collection")
	flag.StringVar(&wrapCmd, "wrap", "", "Wrap a curl command and add it to a collection")
	flag.StringVar(&openapiFile, "openapi", "", "Open an OpenAPI file as a temporary collection")
}

func main() {
	flag.Parse()

	appConfig := config.NewAppConfig()
	store := storage.NewStorage(appConfig.WorkingDir)

	// Handle -i/--init flag
	if initFlag {
		if err := handleInit(store); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if .curlman is initialized (except for OpenAPI mode which uses temp dir)
	if openapiFile == "" && !store.IsInitialized() {
		fmt.Fprintln(os.Stderr, "Error: .curlman directory not found. Run 'curlman -i' to initialize.")
		os.Exit(1)
	}

	// Handle -s/--wrap flag
	if wrapCmd != "" {
		if err := handleWrap(store, wrapCmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle --openapi flag
	if openapiFile != "" {
		if err := handleOpenAPI(openapiFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Default: launch interactive UI
	if err := launchUI(store, appConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleInit(store *storage.Storage) error {
	if store.IsInitialized() {
		fmt.Println(".curlman directory already exists")
		return nil
	}

	if err := store.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	fmt.Println("✓ Initialized .curlman directory")
	fmt.Println("✓ Created default 'None' environment")
	fmt.Println("\nYou can now use curlman to manage your API requests!")
	return nil
}

func handleWrap(store *storage.Storage, curlCmd string) error {
	// Parse and validate curl command
	curlParser := parser.NewCurlParser()
	if err := curlParser.Validate(curlCmd); err != nil {
		return fmt.Errorf("invalid curl command: %w", err)
	}

	// Launch UI in "wrap mode" to select collection
	fmt.Println("Opening interactive view to save curl command...")
	// This would launch the UI with a special mode to save the curl command
	// For now, we'll return an error indicating this needs UI implementation
	return fmt.Errorf("wrap mode requires UI implementation (coming in next step)")
}

func handleOpenAPI(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Get absolute path
	fullPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Parse OpenAPI file (metadata only, spec requests will be loaded on-demand)
	openapiParser := openapi.NewOpenAPIParser()
	collection, err := openapiParser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI file: %w", err)
	}

	// Store the absolute path to the OpenAPI file
	collection.OpenAPIPath = fullPath

	fmt.Printf("✓ Loaded OpenAPI collection: %s\n", collection.Name)
	fmt.Println("\nOpening interactive view to save collection...")

	// Initialize storage if not already done
	appConfig := config.NewAppConfig()
	store := storage.NewStorage(appConfig.WorkingDir)

	if !store.IsInitialized() {
		if err := store.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}
		fmt.Println("✓ Initialized .curlman directory")
	}

	// Save the collection with OpenAPI reference (no requests saved)
	// Spec requests will be generated on-demand when collection is selected
	if err := store.SaveCollection(collection); err != nil {
		return fmt.Errorf("failed to save collection: %w", err)
	}

	fmt.Printf("✓ Saved collection '%s' (spec requests will be loaded on-demand)\n", collection.Name)
	fmt.Println("\nLaunching UI...")

	// Launch UI
	return launchUI(store, appConfig)
}

func launchUI(store *storage.Storage, appConfig *config.AppConfig) error {
	// Launch the interactive TUI
	model := ui.NewModel(store, appConfig)
	return ui.Run(model)
}
