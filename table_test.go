package dynamo

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func TestAddConsumedCapacity(t *testing.T) {
	raw := &types.ConsumedCapacity{
		TableName: aws.String("TestTable"),
		Table: &types.Capacity{
			CapacityUnits:      aws.Float64(9),
			ReadCapacityUnits:  aws.Float64(4),
			WriteCapacityUnits: aws.Float64(5),
		},
		GlobalSecondaryIndexes: map[string]types.Capacity{
			"TestGSI": {
				CapacityUnits:      aws.Float64(3),
				ReadCapacityUnits:  aws.Float64(1),
				WriteCapacityUnits: aws.Float64(2),
			},
		},
		LocalSecondaryIndexes: map[string]types.Capacity{
			"TestLSI": {
				CapacityUnits:      aws.Float64(30),
				ReadCapacityUnits:  aws.Float64(10),
				WriteCapacityUnits: aws.Float64(20),
			},
		},
		CapacityUnits:      aws.Float64(42),
		ReadCapacityUnits:  aws.Float64(15),
		WriteCapacityUnits: aws.Float64(27),
	}
	expected := &ConsumedCapacity{
		TableName:  *raw.TableName,
		Table:      *raw.Table.CapacityUnits,
		TableRead:  *raw.Table.ReadCapacityUnits,
		TableWrite: *raw.Table.WriteCapacityUnits,
		GSI:        map[string]float64{"TestGSI": *raw.GlobalSecondaryIndexes["TestGSI"].CapacityUnits},
		GSIRead:    map[string]float64{"TestGSI": *raw.GlobalSecondaryIndexes["TestGSI"].ReadCapacityUnits},
		GSIWrite:   map[string]float64{"TestGSI": *raw.GlobalSecondaryIndexes["TestGSI"].WriteCapacityUnits},
		LSI:        map[string]float64{"TestLSI": *raw.LocalSecondaryIndexes["TestLSI"].CapacityUnits},
		LSIRead:    map[string]float64{"TestLSI": *raw.LocalSecondaryIndexes["TestLSI"].ReadCapacityUnits},
		LSIWrite:   map[string]float64{"TestLSI": *raw.LocalSecondaryIndexes["TestLSI"].WriteCapacityUnits},
		Total:      *raw.CapacityUnits,
		Read:       *raw.ReadCapacityUnits,
		Write:      *raw.WriteCapacityUnits,
	}

	var cc = new(ConsumedCapacity)
	addConsumedCapacity(cc, raw)

	if !reflect.DeepEqual(cc, expected) {
		t.Error("bad ConsumedCapacity:", cc, "≠", expected)
	}
}
