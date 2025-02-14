package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
)

// ErrDynamoRecordNotFound is used when there is no record in a dynamo table
// (given that GetItem does not actually error)
var ErrDynamoRecordNotFound = errors.New("no record found in dynamo table")

// DynamoProviderContextTable implements the store.ProviderContextTable interface on dynamodb
type DynamoProviderContextTable struct {
	tableName      string
	dynamoDbClient *dynamodb.Client
}

var _ store.ProviderContextTable = (*DynamoProviderContextTable)(nil)

// NewDynamoProviderContextTable returns a ProviderContextTable connected to a AWS DynamoDB table
func NewDynamoProviderContextTable(cfg aws.Config, tableName string, opts ...func(*dynamodb.Options)) *DynamoProviderContextTable {
	return &DynamoProviderContextTable{
		tableName:      tableName,
		dynamoDbClient: dynamodb.NewFromConfig(cfg, opts...),
	}
}

// Delete implements store.ProviderContextTable.
func (d *DynamoProviderContextTable) Delete(ctx context.Context, p peer.ID, contextID []byte) error {
	providerContextItem := providerContextItem{p.String(), contextID, nil}
	_, err := d.dynamoDbClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName), Key: providerContextItem.GetKey(),
	})
	return err
}

// Get implements store.ProviderContextTable.
func (d *DynamoProviderContextTable) Get(ctx context.Context, p peer.ID, contextID []byte) ([]byte, error) {
	providerContextItem := providerContextItem{p.String(), contextID, nil}
	response, err := d.dynamoDbClient.GetItem(ctx, &dynamodb.GetItemInput{
		Key:                  providerContextItem.GetKey(),
		TableName:            aws.String(d.tableName),
		ProjectionExpression: aws.String("contents"),
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving item: %w", err)
	}
	if response.Item == nil {
		return nil, store.NewErrNotFound(ErrDynamoRecordNotFound)
	}
	err = attributevalue.UnmarshalMap(response.Item, &providerContextItem)
	if err != nil {
		return nil, fmt.Errorf("deserializing item: %w", err)
	}
	return providerContextItem.Contents, nil
}

// Put implements store.ProviderContextTable.
func (d *DynamoProviderContextTable) Put(ctx context.Context, p peer.ID, contextID []byte, data []byte) error {
	item, err := attributevalue.MarshalMap(providerContextItem{
		Provider:  p.String(),
		ContextID: contextID,
		Contents:  data,
	})
	if err != nil {
		return fmt.Errorf("serializing item: %w", err)
	}
	_, err = d.dynamoDbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName), Item: item,
	})
	if err != nil {
		return fmt.Errorf("storing item: %w", err)
	}
	return nil
}

type providerContextItem struct {
	Provider  string `dynamodbav:"provider"`
	ContextID []byte `dynamodbav:"contextID"`
	Contents  []byte `dynamodbav:"contents"`
}

// GetKey returns the composite primary key of the provider & contextID in a format that can be
// sent to DynamoDB.
func (p providerContextItem) GetKey() map[string]types.AttributeValue {
	provider, err := attributevalue.Marshal(p.Provider)
	if err != nil {
		panic(err)
	}
	contextID, err := attributevalue.Marshal(p.ContextID)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"provider": provider, "contextID": contextID}
}
