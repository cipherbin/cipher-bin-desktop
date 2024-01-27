package desktop

import (
	"fmt"
	"image/color"
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

// TODO: look into notifications

// API endpoint constants
const (
	WebBaseURL     = "https://cipherb.in"
	APIBaseURL     = "https://api.cipherb.in"
	PrefCurrentTab = "currentTab"
)

// Client defines the main desktop client structure.
type Client struct {
	app            fyne.App
	window         *fyne.Window
	apiClient      *api.Client
	writeInput     *widget.Entry
	readInput      *widget.Entry
	writeForm      *widget.Form
	readForm       *widget.Form
	homeWindow     *fyne.Container
	writeContainer *fyne.Container
	readContainer  *fyne.Container
	tabs           *container.AppTabs
}

// NewClient sets up and initializes a desktop client using the provided http client.
func NewClient(httpClient *http.Client) (*Client, error) {
	c := Client{
		app: app.NewWithID("cipherb.in.desktop"),
	}
	c.app.SetIcon(theme.FyneLogo())
	w := c.app.NewWindow("cipherb.in")
	c.window = &w

	ac, err := api.NewClient(WebBaseURL, APIBaseURL, httpClient)
	if err != nil {
		return nil, err
	}
	c.apiClient = ac

	c.writeInput = widget.NewMultiLineEntry()
	c.readInput = widget.NewMultiLineEntry()
	c.initializeForms()
	c.initializeContainers()
	c.initializeTabs()
	c.initializeWindow()

	return &c, nil
}

// Run runs the desktop application.
func (c *Client) Run() {
	win := *c.window
	win.ShowAndRun()
}

func (c *Client) resetInputs() {
	c.clearInputs()
	c.refreshInputs()
}

func (c *Client) clearInputs() {
	c.writeInput.Text = ""
	c.readInput.Text = ""
}

func (c *Client) refreshInputs() {
	c.writeInput.Refresh()
	c.readInput.Refresh()
}

func (c *Client) writeSubmit() {
	// Ensure we clear and refresh inputs at the end
	defer c.resetInputs()

	uuidv4 := uuid.NewV4().String()
	key := randstring.New(32)

	encryptedMsg, err := aes256.Encrypt([]byte(c.writeInput.Text), key)
	if err != nil {
		// TODO: print to user. Error too?
		fmt.Println("were sorry, there was an error encrypting your message")
		return
	}

	// Create one time use URL with format {host}?bin={uuidv4};{ecryption_key}
	url := fmt.Sprintf("%s/msg?bin=%s;%s", WebBaseURL, uuidv4, key)
	msg := db.Message{UUID: uuidv4, Message: encryptedMsg}

	if err := c.apiClient.PostMessage(&msg); err != nil {
		fmt.Println("were sorry, there was an error sending your message to cipherb.in")
		return
	}

	fmt.Printf("One time URL: %s\n", url)

	urlText := canvas.NewText(url, color.White)
	content := container.New(layout.NewCenterLayout(), urlText)
	c.writeContainer.Add(content)
	// c.WriteContainer.Remove(content)
}

func (c *Client) readSubmit() {
	defer c.resetInputs()

	url := c.readInput.Text
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

	encryptedMsg, err := c.apiClient.GetMessage(apiURL)
	if err != nil {
		fmt.Printf("error: failed to fetch message: %+v", err)
		return
	}

	// Ensure we have what looks like an AES key and set the key var if so
	// Set key to whatever the user has provided for the AES key.
	key := urlParts[1]

	// Length of urlParts != 2. In other words, if it's an invalid link.
	if key == "" {
		// TODO: print to user
		fmt.Println("error: it seems you have an invalid link")
		return
	}

	plainTextMsg, err := aes256.Decrypt(encryptedMsg.Message, key)
	if err != nil {
		// TODO: print to user
		fmt.Printf("error: we had trouble decrypting your message: %+v", err)
		return
	}
	fmt.Println(plainTextMsg)

	text1 := canvas.NewText(plainTextMsg, color.White)
	content := container.New(layout.NewCenterLayout(), text1)
	c.readContainer.Add(content)
	// c.ReadContainer.Remove(content)
}

func (c *Client) initializeForms() {
	c.writeForm = &widget.Form{
		Items:    []*widget.FormItem{{Text: "Message", Widget: c.writeInput}},
		OnCancel: c.resetInputs,
		OnSubmit: c.writeSubmit,
	}
	c.readForm = &widget.Form{
		Items:    []*widget.FormItem{{Text: "URL", Widget: c.readInput}},
		OnCancel: c.resetInputs,
		OnSubmit: c.readSubmit,
	}
}

func (c *Client) initializeContainers() {
	c.initializeHomeContainer()
	c.initializeWriteContainer()
	c.initializeReadContainer()
}

func (c *Client) initializeWriteContainer() {
	c.writeContainer = container.New(
		layout.NewBorderLayout(widget.NewToolbar(), nil, nil, nil),
		container.NewAppTabs(container.NewTabItem("Message", c.writeForm)),
	)
}

func (c *Client) initializeReadContainer() {
	c.readContainer = container.New(
		layout.NewBorderLayout(widget.NewToolbar(), nil, nil, nil),
		container.NewAppTabs(container.NewTabItem("Message", c.readForm)),
	)
}

func (c *Client) initializeHomeContainer() {
	logo := canvas.NewImageFromResource(data.FyneLogo)
	logo.SetMinSize(fyne.NewSize(228, 167))

	c.homeWindow = container.NewVBox(
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
				widget.NewButton("Dark", func() { c.app.Settings().SetTheme(theme.DarkTheme()) }),
				widget.NewButton("Light", func() { c.app.Settings().SetTheme(theme.LightTheme()) }),
			),
		),
	)
}

func (c *Client) initializeTabs() {
	c.tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("Welcome", theme.HomeIcon(), c.homeWindow),
		container.NewTabItemWithIcon("Write Message", theme.DocumentCreateIcon(), c.writeContainer),
		container.NewTabItemWithIcon("Read Message", theme.FolderOpenIcon(), c.readContainer),
	)
	c.tabs.SetTabLocation(container.TabLocationLeading)
	c.tabs.SelectIndex(c.app.Preferences().Int(PrefCurrentTab))
	c.tabs.OnSelected = func(tab *container.TabItem) { c.resetInputs() }
}

func (c *Client) initializeWindow() {
	win := *c.window
	win.SetContent(c.tabs)
	win.ShowAndRun()
	c.app.Preferences().SetInt(PrefCurrentTab, c.tabs.SelectedIndex())
	win.SetContent(
		container.New(
			layout.NewBorderLayout(widget.NewToolbar(), nil, nil, nil),
			c.tabs,
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
