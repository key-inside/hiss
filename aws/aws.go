package aws

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/viper"
)

type configOption struct {
	baseEndpoint *string
	region       *string
}

type Option func(o *configOption)

func WithEndpoint(url string) Option {
	return func(o *configOption) {
		o.baseEndpoint = aws.String(url)
	}
}

func WithRegion(region string) Option {
	return func(o *configOption) {
		o.region = aws.String(region)
	}
}

func resolveConfigOption(ops ...Option) *configOption {
	o := &configOption{}
	for _, op := range ops {
		op(o)
	}
	return o
}

func ext(resource string) string {
	for i := len(resource) - 1; i >= 0 && resource[i] != '/' && resource[i] != ':'; i-- {
		if resource[i] == '.' {
			return resource[i+1:]
		}
	}
	return ""
}

func GetConfigMap(arnStr string, ops ...Option) (map[string]any, error) {
	str, err := GetString(arnStr)
	if err != nil {
		return nil, err
	}
	v := viper.New()
	v.SetConfigType(ext(arnStr))
	if err := v.ReadConfig(strings.NewReader(str)); err != nil {
		return nil, fmt.Errorf("failed to read from string reader: %w", err)
	}
	return v.AllSettings(), nil
}

func GetString(arnStr string, ops ...Option) (string, error) {
	a, err := arn.Parse(arnStr)
	if err != nil {
		return "", err
	}
	switch a.Service {
	case "secretsmanager":
		return GetStringFromSecretsManager(a, ops...)
	case "ssm":
		return GetStringFromParameterStore(a, ops...)
	}
	return "", fmt.Errorf("not supported AWS service: %s", a.Service)
}

func defaultConfig(o *configOption) (aws.Config, error) {
	if o.region != nil {
		return config.LoadDefaultConfig(context.TODO(), config.WithRegion(*o.region))
	}
	return config.LoadDefaultConfig(context.TODO())
}

func NewClient[O any, C any](newFromConfig func(aws.Config, ...func(*O)) *C, ops ...Option) (*C, error) {
	cfgOps := resolveConfigOption(ops...)

	cfg, err := defaultConfig(cfgOps)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	optFns := []func(*O){}
	if cfgOps.baseEndpoint != nil {
		optFns = append(optFns, func(o *O) {
			reflect.ValueOf(o).Elem().FieldByName("BaseEndpoint").Set(reflect.ValueOf(cfgOps.baseEndpoint))
		})
	}

	return newFromConfig(cfg, optFns...), nil
}

func GetStringFromSecretsManager(a arn.ARN, ops ...Option) (string, error) {
	client, err := NewClient(secretsmanager.NewFromConfig, ops...)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	res, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(strings.TrimPrefix(a.Resource, "secret:")),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret value: %w", err)
	}

	return *res.SecretString, nil
}

func GetStringFromParameterStore(a arn.ARN, ops ...Option) (string, error) {
	client, err := NewClient(ssm.NewFromConfig, ops...)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	res, err := client.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name: aws.String(strings.TrimPrefix(a.Resource, "parameter")),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get parameter: %w", err)
	}

	return *res.Parameter.Value, nil
}

func GetItemFromDynamoDB(tableName string, key map[string]any, ops ...Option) (map[string]types.AttributeValue, error) {
	keyMap := map[string]types.AttributeValue{}
	var err error
	for k, v := range key {
		keyMap[k], err = attributevalue.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal key: %w", err)
		}
	}

	client, err := NewClient(dynamodb.NewFromConfig, ops...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	res, err := client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       keyMap,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return res.Item, nil
}

func PutItemToDynamoDB(tableName string, item any, ops ...Option) error {
	itemMap, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	client, err := NewClient(dynamodb.NewFromConfig, ops...)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	_, err = client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      itemMap,
	})
	return err
}
