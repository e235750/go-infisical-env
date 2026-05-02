package infisicalenv

// Config はクライアントを初期化するための設定項目を保持します。
type Config struct {
	SiteURL      string
	ClientID     string
	ClientSecret string
	Environment  string
	ProjectID    string
	SecretPath   string // デフォルトは "/"
}