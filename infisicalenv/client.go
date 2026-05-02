package infisicalenv

import (
	"context"
	"fmt"
	"os"
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

	return c.loadRecursive(rv.Elem(), "")
}

func (c *Client) loadRecursive(val reflect.Value, prefix string) error {
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		fieldInfo := typ.Field(i)
		fieldVal := val.Field(i)

		tag := fieldInfo.Tag.Get("infisical")

		// 構造体の場合は再帰的に処理
		if fieldVal.Kind() == reflect.Struct {
			newPrefix := prefix
			if tag != "" {
				if newPrefix != "" {
					newPrefix = newPrefix + "_" + tag
				} else {
					newPrefix = tag
				}
			}
			if err := c.loadRecursive(fieldVal, newPrefix); err != nil {
				return err
			}
			continue
		}

		if tag == "" {
			continue
		}

		// 接頭辞がある場合は結合する
		fullKey := tag
		if prefix != "" {
			fullKey = prefix + "_" + tag
		}

		secretVal, ok := c.secretMap[fullKey]
		if !ok || secretVal == "" {
			secretVal = os.Getenv(fullKey)
		}

		// 値が存在する場合のみセット
		if secretVal != "" && fieldVal.CanSet() && fieldVal.Kind() == reflect.String {
			fieldVal.SetString(secretVal)
		}
	}
	return nil
}

