package infisicalenv

import (
	"context"
	"fmt"
	"reflect"

	infisical "github.com/infisical/go-sdk"
)

// Client はInfisicalのSDKをラップし、
// 環境変数とInfisicalのシークレットを統合して取得するための構造体です。
type Client struct {
	secretMap map[string]string
}

// NewClient は新しいInfisicalラッパークライアントを作成します。
func NewClient(config Config) (*Client, error) {
	ctx := context.Background()
	client := infisical.NewInfisicalClient(ctx, infisical.Config{
		SiteUrl: config.SiteURL,
	})

	_, err := client.Auth().UniversalAuthLogin(
		config.ClientID,
		config.ClientSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with infisical: %w", err)
	}

	list, err := client.Secrets().ListSecrets(
		infisical.ListSecretsOptions{
			Environment: config.Environment,
			ProjectID:   config.ProjectID,
			SecretPath:  config.SecretPath,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	secretMap := make(map[string]string)
	for _, s := range list.Secrets {
		secretMap[s.SecretKey] = s.SecretValue
	}

	return &Client{secretMap: secretMap}, nil
}

func (c *Client) LoadConfig(ptr any) error {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("ptr must be a pointer to a struct")
	}

	val := rv.Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		fieldInfo := typ.Field(i)
		fieldVal := val.Field(i)

		tag := fieldInfo.Tag.Get("infisical")
		if tag == "" {
			continue
		}

		secretVal := c.secretMap[tag]

		// 値が存在する場合のみセット（空文字で上書きしないようにする）
		if secretVal != "" && fieldVal.CanSet() && fieldVal.Kind() == reflect.String {
			fieldVal.SetString(secretVal)
		}
	}
	return nil
}

