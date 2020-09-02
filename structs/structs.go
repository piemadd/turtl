package structs

type User struct {
	DiscordID   string
	APIKey      string
	UploadLimit int64
	Admin       bool
}

type DiscordExchangeRequest struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RedirectUri  string `json:"redirect_uri"`
	Scope        string `json:"scope"'`
}

type DiscordExchangeResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type PartialGuild struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	Icon           string   `json:"icon"`
	Owner          bool     `json:"owner"`
	Permissions    int      `json:"permissions"`
	Features       []string `json:"features"`
	PermissionsNew string   `json:"permissions_new"`
}

type Blacklist struct {
	SHA256 string
	Reason string
}

type UploadRequest struct {
	Bucket string `json:"domain"`
	APIKey string `json:"api_key"`
}

type FileUploadResponse struct {
	Success bool   `json:"success"`
	Status  int    `json:"status"`
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	Info    string `json:"info,omitempty"`
}

type Object struct {
	Bucket    string
	Wildcard  string
	FileName  string
	Uploader  string
	CreatedAt int
	MD5       string
	SHA256    string
	DeletedAt int
}
