package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

func Test_AWSParameterStore(t *testing.T) {
	arn := "arn:aws:ssm:::parameter/test/animal/config.yaml"
	str, err := GetString(arn)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		t.Log(str)
	}
}

func Test_AWSSecretsManager(t *testing.T) {
	arn := "arn:aws:secretsmanager:::secret:test/animal/secrets.yaml"
	str, err := GetString(arn)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		t.Log(str)
	}
}

/*
# Local DynamoDB test

1. pull & run docker image `amazon/dynamodb-local`
```sh
% docker pull amazon/dynamodb-local
% docker run -p 8000:8000 amazon/dynamodb-local
```

2. create a test table schema file
```sh
% echo '{
	"TableName": "hiss",
	"AttributeDefinitions": [
		{
			"AttributeName": "animal",
			"AttributeType": "S"
		}
	],
	"KeySchema": [
		{
			"AttributeName": "animal",
			"KeyType": "HASH"
		}
	],
	"BillingMode": "PAY_PER_REQUEST"
}' > create-table.json
```

3. create a table
```sh
% aws dynamodb create-table --endpoint-url http://localhost:8000 --cli-input-json file://create-table.json
```
*/

func Test_AWSPutDynamoDBItem(t *testing.T) {
	item := map[string]any{"animal": "snake", "foot": 0}
	err := PutItemToDynamoDB("hiss", item, WithEndpoint("http://localhost:8000"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func Test_AWSGetDynamoDBItem(t *testing.T) {
	key := map[string]any{"animal": "snake"}
	item, err := GetItemFromDynamoDB("hiss", key, WithEndpoint("http://localhost:8000"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		var foot int
		attributevalue.Unmarshal(item["foot"], &foot)
		t.Log("snake foot =>", foot)
	}
}
