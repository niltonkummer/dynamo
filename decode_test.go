package dynamo

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var itemDecodeOnlyTests = []struct {
	name   string
	given  map[string]types.AttributeValue
	expect interface{}
}{
	{
		// unexported embedded pointers should be ignored
		name: "embedded unexported pointer",
		given: map[string]types.AttributeValue{
			"Embedded": &types.AttributeValueMemberBOOL{Value: true},
		},
		expect: struct {
			*embedded
		}{},
	},
	{
		// unexported fields should be ignored
		name: "unexported fields",
		given: map[string]types.AttributeValue{
			"a": &types.AttributeValueMemberBOOL{Value: true},
		},
		expect: struct {
			a bool
		}{},
	},
	{
		// embedded pointers shouldn't clobber existing fields
		name: "exported pointer embedded struct clobber",
		given: map[string]types.AttributeValue{
			"Embedded": &types.AttributeValueMemberS{Value: "OK"},
		},
		expect: struct {
			Embedded string
			*ExportedEmbedded
		}{
			Embedded:         "OK",
			ExportedEmbedded: &ExportedEmbedded{},
		},
	},
}

func TestUnmarshalAsymmetric(t *testing.T) {
	for _, tc := range itemDecodeOnlyTests {
		rv := reflect.New(reflect.TypeOf(tc.expect))
		expect := rv.Interface()
		err := UnmarshalItem(tc.given, expect)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}
		if !reflect.DeepEqual(rv.Elem().Interface(), tc.expect) {
			t.Errorf("%s: bad result: %#v ≠ %#v", tc.name, rv.Elem().Interface(), tc.expect)
		}
	}
}

func TestUnmarshalAppend(t *testing.T) {
	var results []struct {
		User  int `dynamo:"UserID"`
		Page  int
		Limit uint
		Null  interface{}
	}
	id := "12345"
	page := "5"
	limit := "20"
	null := true
	item := map[string]types.AttributeValue{
		"UserID": &types.AttributeValueMemberN{Value: id},
		"Page":   &types.AttributeValueMemberN{Value: page},
		"Limit":  &types.AttributeValueMemberN{Value: limit},
		"Null":   &types.AttributeValueMemberNULL{Value: null},
	}

	for range [15]struct{}{} {
		err := unmarshalAppend(item, &results)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, h := range results {
		if h.User != 12345 || h.Page != 5 || h.Limit != 20 || h.Null != nil {
			t.Error("invalid hit", h)
		}
	}

	var mapResults []map[string]interface{}

	for range [15]struct{}{} {
		err := unmarshalAppend(item, &mapResults)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, h := range mapResults {
		if h["UserID"] != 12345.0 || h["Page"] != 5.0 || h["Limit"] != 20.0 || h["Null"] != nil {
			t.Error("invalid interface{} hit", h)
		}
	}
}

func TestUnmarshalItem(t *testing.T) {
	for _, tc := range itemEncodingTests {
		rv := reflect.New(reflect.TypeOf(tc.in))
		err := unmarshalItem(tc.out, rv.Interface())
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
		}

		if !reflect.DeepEqual(rv.Elem().Interface(), tc.in) {
			var opts []cmp.Option
			if rv.Elem().Kind() == reflect.Struct {
				opts = append(opts, cmpopts.IgnoreUnexported(rv.Elem().Interface()))
			}

			diff := cmp.Diff(rv.Elem().Interface(), tc.in, opts...)
			fmt.Println(diff)

			if strings.TrimSpace(diff) != "" {
				t.Errorf("%s: bad result: %#v ≠ %#v", tc.name, rv.Elem().Interface(), tc.in)
			}
		}

	}
}
