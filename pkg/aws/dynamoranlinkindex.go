package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

// DynamoRanLinkIndex implements the store.ProviderContextTable interface on dynamodb
type DynamoRanLinkIndex struct {
	tableName      string
	dynamoDbClient *dynamodb.Client
}

var _ receiptstore.RanLinkIndex = (*DynamoRanLinkIndex)(nil)

// NewDynamoRanLinkIndex returns a ProviderContextTable connected to a AWS DynamoDB table
func NewDynamoRanLinkIndex(cfg aws.Config, tableName string, opts ...func(*dynamodb.Options)) *DynamoRanLinkIndex {
	return &DynamoRanLinkIndex{
		tableName:      tableName,
		dynamoDbClient: dynamodb.NewFromConfig(cfg, opts...),
	}
}

// Get implements store.ProviderContextTable.
func (d *DynamoRanLinkIndex) Get(ctx context.Context, ran datamodel.Link) (datamodel.Link, error) {
	ranLinkItem := ranLinkItem{Ran: ran.String()}
	response, err := d.dynamoDbClient.GetItem(ctx, &dynamodb.GetItemInput{
		Key:                  ranLinkItem.GetKey(),
		TableName:            aws.String(d.tableName),
		ProjectionExpression: aws.String("link"),
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving item: %w", err)
	}
	if response.Item == nil {
		return nil, store.NewErrNotFound(ErrDynamoRecordNotFound)
	}
	err = attributevalue.UnmarshalMap(response.Item, &ranLinkItem)
	if err != nil {
		return nil, fmt.Errorf("deserializing item: %w", err)
	}
	cid, err := cid.Decode(ranLinkItem.Link)
	if err != nil {
		return nil, fmt.Errorf("decoding link: %w", err)
	}
	return cidlink.Link{Cid: cid}, nil
}

// Put implements store.ProviderContextTable.
func (d *DynamoRanLinkIndex) Put(ctx context.Context, ran datamodel.Link, link datamodel.Link) error {
	item, err := attributevalue.MarshalMap(ranLinkItem{
		Ran:  ran.String(),
		Link: link.String(),
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

type ranLinkItem struct {
	Ran  string `dynamodbav:"ran"`
	Link string `dynamodbav:"link"`
}

// GetKey returns the composite primary key of the provider & contextID in a format that can be
// sent to DynamoDB.
func (p ranLinkItem) GetKey() map[string]types.AttributeValue {
	ran, err := attributevalue.Marshal(p.Ran)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"ran": ran}
}
