package desktop

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_demo/data"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cipherbin/cipher-bin-cli/pkg/aes256"
	"github.com/cipherbin/cipher-bin-cli/pkg/api"
	"github.com/cipherbin/cipher-bin-cli/pkg/randstring"
	"github.com/cipherbin/cipher-bin-server/db"
	uuid "github.com/satori/go.uuid"
)

// API endpoint constants
const (
	WebBaseURL     = "https://cipherb.in"
	APIBaseURL     = "https://api.cipherb.in"
	PrefCurrentTab = "currentTab"
)

// Client ...
type Client struct {
	App            fyne.App
	Window         *fyne.Window
	APIClient      *api.Client
	WriteInput     *widget.Entry
	ReadInput      *widget.Entry
	WriteForm      *widget.Form
	ReadForm       *widget.Form
	HomeWindow     *fyne.Container
	WriteContainer *fyne.Container
	ReadContainer  *fyne.Container
	Tabs           *container.AppTabs
}

// NewClient ...
func NewClient(httpClient *http.Client) (*Client, error) {
	dc := new(Client)
	dc.App = app.NewWithID("cipherb.in.desktop")
	dc.App.SetIcon(theme.FyneLogo())
	w := dc.App.NewWindow("cipherb.in")
	dc.Window = &w

	ac, err := api.NewClient(WebBaseURL, APIBaseURL, httpClient)
	if err != nil {
		return nil, err
	}
	dc.APIClient = ac

	dc.WriteInput = widget.NewMultiLineEntry()
	dc.ReadInput = widget.NewMultiLineEntry()
	dc.initializeForms()
	dc.initializeContainers()
	dc.initializeTabs()
	dc.initializeWindow()

	return dc, nil
}

// Run ...
func (c *Client) Run() {
	win := *c.Window
	win.ShowAndRun()
}

// resetInputs ...
func (c *Client) resetInputs() {
	c.clearInputs()
	c.refreshInputs()
}

func (c *Client) clearInputs() {
	c.WriteInput.Text = ""
	c.ReadInput.Text = ""
}

func (c *Client) refreshInputs() {
	c.WriteInput.Refresh()
	c.ReadInput.Refresh()
}

// writeSubmit ...
func (c *Client) writeSubmit() {
	// Ensure we clear and refresh inputs at the end
	defer c.resetInputs()

	uuidv4 := uuid.NewV4().String()
	key := randstring.New(32)

	encryptedMsg, err := aes256.Encrypt([]byte(c.WriteInput.Text), key)
	if err != nil {
		// TODO: print to user. Error too?
		fmt.Println("were sorry, there was an error encrypting your message")
		return
	}

	// Create one time use URL with format {host}?bin={uuidv4};{ecryption_key}
	url := fmt.Sprintf("%s/msg?bin=%s;%s", WebBaseURL, uuidv4, key)
	msg := db.Message{UUID: uuidv4, Message: encryptedMsg}

	if err := c.APIClient.PostMessage(&msg); err != nil {
		fmt.Println("were sorry, there was an error sending your message to cipherb.in")
		return
	}

	fmt.Printf("One time URL: %s\n", url)
	// TODO: write to screen
}

// readSubmit ...
func (c *Client) readSubmit() {
	defer c.resetInputs()

	url := c.ReadInput.Text
	if !validURL(url, WebBaseURL) {
		// TODO: print to user
		fmt.Println("sorry, this message has either already been viewed and destroyed or it never existed at all")
		return
	}

	// Replace the browser url with the api url to fetch the message.
	url = strings.Replace(url, WebBaseURL, APIBaseURL, -1)
	urlParts := strings.Split(url, ";")
	if len(urlParts) != 2 {
		// TODO: print to user
		fmt.Println("Sorry, that seems to be an invalid cipherbin link")
	}
	apiURL := urlParts[0] // uuid only

	encryptedMsg, err := c.APIClient.GetMessage(apiURL)
	if err != nil {
		fmt.Printf("error: failed to fetch message: %+v", err)
		return
	}

	// Ensure we have what looks like an AES key and set the key var if so
	// Set key to whatever the user has provided for the AES key.
	key := urlParts[1]

	// Length of urlParts != 2. In other words, if it's an invalid link.
	if key == "" {
		fmt.Printf("error: it seems you have an invalid link: %+v", err)
		return
	}

	plainTextMsg, err := aes256.Decrypt(encryptedMsg.Message, key)
	if err != nil {
		// TODO: print to user
		fmt.Printf("error: we had trouble decrypting your message: %+v", err)
		return
	}
	fmt.Println(plainTextMsg)
	// TODO: write to screen
}

// initializeForms ...
func (c *Client) initializeForms() {
	c.WriteForm = &widget.Form{
		Items:    []*widget.FormItem{{Text: "Message", Widget: c.WriteInput}},
		OnCancel: c.resetInputs,
		OnSubmit: c.writeSubmit,
	}
	c.ReadForm = &widget.Form{
		Items:    []*widget.FormItem{{Text: "URL", Widget: c.ReadInput}},
		OnCancel: c.resetInputs,
		OnSubmit: c.readSubmit,
	}
}

// initializeContainers ...
func (c *Client) initializeContainers() {
	c.initializeHomeContainer()
	c.initializeWriteContainer()
	c.initializeReadContainer()
}

func (c *Client) initializeWriteContainer() {
	c.WriteContainer = container.New(
		layout.NewBorderLayout(widget.NewToolbar(), nil, nil, nil),
		container.NewAppTabs(container.NewTabItem("Message", c.WriteForm)),
	)
}

func (c *Client) initializeReadContainer() {
	c.ReadContainer = container.New(
		layout.NewBorderLayout(widget.NewToolbar(), nil, nil, nil),
		container.NewAppTabs(container.NewTabItem("Message", c.ReadForm)),
	)
}

func (c *Client) initializeHomeContainer() {
	logo := canvas.NewImageFromResource(data.FyneLogo)
	logo.SetMinSize(fyne.NewSize(228, 167))

	c.HomeWindow = container.NewVBox(
		layout.NewSpacer(),
		widget.NewLabelWithStyle(
			"welcome to cipherb.in",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		container.NewHBox(layout.NewSpacer(), logo, layout.NewSpacer()),
		container.NewHBox(
			layout.NewSpacer(),
			widget.NewHyperlink("cipherb.in", parseURL("https://cipherb.in/")),
			widget.NewLabel("-"),
			widget.NewHyperlink("github", parseURL("https://github.com/cipherbin/cipher-bin-desktop")),
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
		container.NewHBox(
			widget.NewLabelWithStyle(
				"Theme",
				fyne.TextAlignCenter,
				fyne.TextStyle{Bold: true},
			),
			container.New(
				layout.NewGridLayout(2),
				widget.NewButton("Dark", func() { c.App.Settings().SetTheme(theme.DarkTheme()) }),
				widget.NewButton("Light", func() { c.App.Settings().SetTheme(theme.LightTheme()) }),
			),
		),
	)
}

// initializeTabs ...
func (c *Client) initializeTabs() {
	c.Tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("Welcome", theme.HomeIcon(), c.HomeWindow),
		container.NewTabItemWithIcon("Write Message", theme.DocumentCreateIcon(), c.WriteContainer),
		container.NewTabItemWithIcon("Read Message", theme.FolderOpenIcon(), c.ReadContainer),
	)
	c.Tabs.SetTabLocation(container.TabLocationLeading)
	c.Tabs.SelectIndex(c.App.Preferences().Int(PrefCurrentTab))
	c.Tabs.OnSelected = func(tab *container.TabItem) { c.resetInputs() }
}

// initializeWindow ...
func (c *Client) initializeWindow() {
	win := *c.Window
	win.SetContent(c.Tabs)
	win.ShowAndRun()
	c.App.Preferences().SetInt(PrefCurrentTab, c.Tabs.CurrentTabIndex())
	win.SetContent(
		container.New(
			layout.NewBorderLayout(widget.NewToolbar(), nil, nil, nil),
			c.Tabs,
		),
	)
}

// validURL takes a string url and checks whether it's a valid cipherb.in link
func validURL(url, apiBaseURL string) bool {
	return strings.HasPrefix(url, fmt.Sprintf("%s/msg?bin=", apiBaseURL))
}

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}
	return link
}
