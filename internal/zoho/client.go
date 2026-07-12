package zoho

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"X-BE/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrNotConfigured      = errors.New("Zoho is not configured")
	ErrTokenNotFound      = errors.New("Zoho token not found in database")
	ErrMailNotConfigured  = errors.New("Zoho mail is not configured")
	ErrDriveNotConfigured = errors.New("Zoho drive is not configured")
	ErrSheetNotConfigured = errors.New("Zoho sheet is not configured")
)

const tokenProvider = "zoho"

const (
	zohoAuthBaseURL   = "https://accounts.zoho.in"
	zohoSheetsBaseURL = "https://sheet.zoho.in"
	defaultZohoScope  = "WorkDrive.files.CREATE,ZohoMail.accounts.READ,ZohoMail.messages.CREATE,ZohoSheet.dataAPI.READ,ZohoSheet.dataAPI.UPDATE,offline_access"
)

var (
	zohoMailEndpoints = []string{
		"https://mail.zoho.in/api/accounts/%s/messages",
		"https://mail.zoho.com/api/accounts/%s/messages",
	}
	zohoDriveUploadEndpoints = []string{
		"https://www.zohoapis.in/workdrive/api/v1/files",
		"https://workdrive.zoho.in/api/v1/files",
		"https://www.zohoapis.in/workdrive/api/v1/upload",
		"https://workdrive.zoho.in/api/v1/upload",
	}
)

type tokenDocument struct {
	AccountID     string    `bson:"accountId"`
	Provider      string    `bson:"provider"`
	AccessToken   string    `bson:"accessToken"`
	RefreshToken  string    `bson:"refreshToken"`
	ExpiresAt     time.Time `bson:"expiresAt"`
	LastRefreshed time.Time `bson:"lastRefreshed"`
}

type EmailRequest struct {
	AccountID   string
	FromAddress string
	To          []string
	Subject     string
	Content     string
	ContentType string
}

type DriveUploadRequest struct {
	ParentFolderID string
	Filename       string
	MimeType       string
	Content        []byte
}

type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
}

type Client struct {
	httpClient   *http.Client
	cfg          config.ZohoConfig
	tokens       *mongo.Collection
	mu           sync.Mutex
	accessToken  string
	refreshToken string
	accountID    string
	expiresAt    time.Time
}

func NewClient(cfg config.Config, db *mongo.Database) *Client {
	timeout := cfg.Zoho.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	var tokenCollection *mongo.Collection
	if db != nil {
		tokenCollection = db.Collection("zoho_tokens")
	}
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		cfg:        cfg.Zoho,
		tokens:     tokenCollection,
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.cfg.ClientID != "" && c.cfg.ClientSecret != "" && c.tokens != nil
}
