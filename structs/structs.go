package structs

type User struct {
	DiscordID		string
	APIKey			string
	Admin			bool
}

type UploadRequest struct {
	Bucket			string `json:"domain"`
	APIKey			string `json:"api_key"`
}

type FileUploadResponse struct {
	Success			bool 	`json:"success"`
	Status			int	 	`json:"status"`
	Name			string	`json:"name,omitempty"`
	URL				string	`json:"url,omitempty"`
	Info			string	`json:"info,omitempty"`
}

type Object struct {
	Bucket			string
	Wildcard		string
	FileName		string
	Uploader		string
	CreatedAt		int
	MD5				string
	SHA256			string
}