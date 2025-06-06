package options

import (
	"context"
	"html"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/wailsapp/wails/v2/pkg/menu"

	"github.com/wailsapp/wails/v2/pkg/logger"
)

type WindowStartState int

const (
	Normal     WindowStartState = 0
	Maximised  WindowStartState = 1
	Minimised  WindowStartState = 2
	Fullscreen WindowStartState = 3
)

type Experimental struct{}

// App contains options for creating the App
type App struct {
	Title             string
	Width             int
	Height            int
	DisableResize     bool
	Fullscreen        bool
	Frameless         bool
	MinWidth          int
	MinHeight         int
	MaxWidth          int
	MaxHeight         int
	StartHidden       bool
	HideWindowOnClose bool
	AlwaysOnTop       bool
	// BackgroundColour is the background colour of the window
	// You can use the options.NewRGB and options.NewRGBA functions to create a new colour
	BackgroundColour *RGBA
	// Deprecated: Use AssetServer.Assets instead.
	Assets fs.FS
	// Deprecated: Use AssetServer.Handler instead.
	AssetsHandler http.Handler
	// AssetServer configures the Assets for the application
	AssetServer        *assetserver.Options
	Menu               *menu.Menu
	Logger             logger.Logger `json:"-"`
	LogLevel           logger.LogLevel
	LogLevelProduction logger.LogLevel
	OnStartup          func(ctx context.Context)                `json:"-"`
	OnDomReady         func(ctx context.Context)                `json:"-"`
	OnShutdown         func(ctx context.Context)                `json:"-"`
	OnBeforeClose      func(ctx context.Context) (prevent bool) `json:"-"`
	Bind               []interface{}
	EnumBind           []interface{}
	WindowStartState   WindowStartState

	// ErrorFormatter overrides the formatting of errors returned by backend methods
	ErrorFormatter ErrorFormatter

	// CSS property to test for draggable elements. Default "--wails-draggable"
	CSSDragProperty string

	// The CSS Value that the CSSDragProperty must have to be draggable, EG: "drag"
	CSSDragValue string

	// EnableDefaultContextMenu enables the browser's default context-menu in production
	// This menu is already enabled in development and debug builds
	EnableDefaultContextMenu bool

	// EnableFraudulentWebsiteDetection enables scan services for fraudulent content, such as malware or phishing attempts.
	// These services might send information from your app like URLs navigated to and possibly other content to cloud
	// services of Apple and Microsoft.
	EnableFraudulentWebsiteDetection bool

	SingleInstanceLock *SingleInstanceLock

	Windows *windows.Options
	Mac     *mac.Options
	Linux   *linux.Options

	// Experimental options
	Experimental *Experimental

	// Debug options for debug builds. These options will be ignored in a production build.
	Debug Debug

	// DragAndDrop options for drag and drop behavior
	DragAndDrop *DragAndDrop

	// DisablePanicRecovery disables the panic recovery system in messages processing
	DisablePanicRecovery bool
}

type ErrorFormatter func(error) any

type RGBA struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

// NewRGBA creates a new RGBA struct with the given values
func NewRGBA(r, g, b, a uint8) *RGBA {
	return &RGBA{
		R: r,
		G: g,
		B: b,
		A: a,
	}
}

// NewRGB creates a new RGBA struct with the given values and Alpha set to 255
func NewRGB(r, g, b uint8) *RGBA {
	return &RGBA{
		R: r,
		G: g,
		B: b,
		A: 255,
	}
}

// MergeDefaults will set the minimum default values for an application
func MergeDefaults(appoptions *App) {
	// Do set defaults
	if appoptions.Width <= 0 {
		appoptions.Width = 1024
	}
	if appoptions.Height <= 0 {
		appoptions.Height = 768
	}
	if appoptions.Logger == nil {
		appoptions.Logger = logger.NewDefaultLogger()
	}
	if appoptions.LogLevel == 0 {
		appoptions.LogLevel = logger.INFO
	}
	if appoptions.LogLevelProduction == 0 {
		appoptions.LogLevelProduction = logger.ERROR
	}
	if appoptions.CSSDragProperty == "" {
		appoptions.CSSDragProperty = "--wails-draggable"
	}
	if appoptions.CSSDragValue == "" {
		appoptions.CSSDragValue = "drag"
	}
	if appoptions.DragAndDrop == nil {
		appoptions.DragAndDrop = &DragAndDrop{}
	}
	if appoptions.DragAndDrop.CSSDropProperty == "" {
		appoptions.DragAndDrop.CSSDropProperty = "--wails-drop-target"
	}
	if appoptions.DragAndDrop.CSSDropValue == "" {
		appoptions.DragAndDrop.CSSDropValue = "drop"
	}
	if appoptions.BackgroundColour == nil {
		appoptions.BackgroundColour = &RGBA{
			R: 255,
			G: 255,
			B: 255,
			A: 255,
		}
	}

	// Ensure max and min are valid
	processMinMaxConstraints(appoptions)

	// Default menus
	processMenus(appoptions)

	// Process Drag Options
	processDragOptions(appoptions)
}

type SingleInstanceLock struct {
	// uniqueId that will be used for setting up messaging between instances
	UniqueId               string
	OnSecondInstanceLaunch func(secondInstanceData SecondInstanceData)
}

type SecondInstanceData struct {
	Args             []string
	WorkingDirectory string
}

type DragAndDrop struct {

	// EnableFileDrop enables wails' drag and drop functionality that returns the dropped in files' absolute paths.
	EnableFileDrop bool

	// Disable webview's drag and drop functionality.
	//
	// It can be used to prevent accidental file opening of dragged in files in the webview, when there is no need for drag and drop.
	DisableWebViewDrop bool

	// CSS property to test for drag and drop target elements. Default "--wails-drop-target"
	CSSDropProperty string

	// The CSS Value that the CSSDropProperty must have to be a valid drop target. Default "drop"
	CSSDropValue string
}

func NewSecondInstanceData() (*SecondInstanceData, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	workingDirectory := filepath.Dir(ex)

	return &SecondInstanceData{
		Args:             os.Args[1:],
		WorkingDirectory: workingDirectory,
	}, nil
}

func processMenus(appoptions *App) {
	switch runtime.GOOS {
	case "darwin":
		if appoptions.Menu == nil {
			items := []*menu.MenuItem{
				menu.EditMenu(),
			}
			if !appoptions.Frameless {
				items = append(items, menu.WindowMenu()) // Current options in Window Menu only work if not frameless
			}

			appoptions.Menu = menu.NewMenuFromItems(menu.AppMenu(), items...)
		}
	}
}

func processMinMaxConstraints(appoptions *App) {
	if appoptions.MinWidth > 0 && appoptions.MaxWidth > 0 {
		if appoptions.MinWidth > appoptions.MaxWidth {
			appoptions.MinWidth = appoptions.MaxWidth
		}
	}
	if appoptions.MinHeight > 0 && appoptions.MaxHeight > 0 {
		if appoptions.MinHeight > appoptions.MaxHeight {
			appoptions.MinHeight = appoptions.MaxHeight
		}
	}
	// Ensure width and height are limited if max/min is set
	if appoptions.Width < appoptions.MinWidth {
		appoptions.Width = appoptions.MinWidth
	}
	if appoptions.MaxWidth > 0 && appoptions.Width > appoptions.MaxWidth {
		appoptions.Width = appoptions.MaxWidth
	}
	if appoptions.Height < appoptions.MinHeight {
		appoptions.Height = appoptions.MinHeight
	}
	if appoptions.MaxHeight > 0 && appoptions.Height > appoptions.MaxHeight {
		appoptions.Height = appoptions.MaxHeight
	}
}

func processDragOptions(appoptions *App) {
	appoptions.CSSDragProperty = html.EscapeString(appoptions.CSSDragProperty)
	appoptions.CSSDragValue = html.EscapeString(appoptions.CSSDragValue)
}
