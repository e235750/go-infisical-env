package main

import (
	"log"
	"os"

	"github.com/e235750/go-infisical-env/infisicalenv"
)

func main() {
	// 環境変数はアプリケーション側で責任を持って読み込む (.envのロード等はここで行う)
	// _ = godotenv.Load()

	cfg := infisicalenv.Config{
		SiteURL:      os.Getenv("INFISICAL_SITE_URL"),
		ClientID:     os.Getenv("INFISICAL_CLIENT_ID"),
		ClientSecret: os.Getenv("INFISICAL_CLIENT_SECRET"),
		Environment:  os.Getenv("INFISICAL_ENVIRONMENT"),
		ProjectID:    os.Getenv("INFISICAL_PROJECT_ID"),
		SecretPath:   "/",
	}

	// クライアントの初期化（内部でInfisicalと通信し、シークレットを取得・保持）
	client, err := infisicalenv.NewClient(cfg)
	if err != nil {
		log.Fatalf("Infisicalクライアントの初期化に失敗しました: %v", err)
	}

	// 例: アプリケーション側の設定構造体を定義
	type AppConfig struct {
		PORT               string `infisical:"PORT"`
		DATABASE_URL        string `infisical:"DATABASE_URL"`
	}

	// 取得したシークレット（または環境変数）を設定構造体にマッピング
	var appCfg AppConfig
	if err := client.LoadConfig(&appCfg); err != nil {
		log.Fatalf("設定のマッピングに失敗しました: %v", err)
	}
}
