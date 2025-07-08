package jpack

import (
	"context"
	"reflect"
	"testing"
)

type mockField struct {
	name      string
	fieldType JFieldType
	schema    JSchema
}

func (f *mockField) Name() string {
	return f.name
}

func (f *mockField) Type() JFieldType {
	return f.fieldType
}

func (f *mockField) Schema() JSchema {
	return f.schema
}
func (f *mockField) Default() any {
	return nil
}

func (f *mockField) Validate(value any) error {
	return nil
}

func TestNumber_Scan(t *testing.T) {

	num := int(42)
	type args struct {
		ctx   context.Context
		field JField
		row   map[string]any
	}
	tests := []struct {
		name      string
		n         *Number
		args      args
		wantValue any
		wantErr   bool
	}{
		{
			name: "Valid integer",
			n:    &Number{},
			args: args{
				ctx:   context.Background(),
				field: &mockField{name: "testField", fieldType: &Number{}},
				row:   map[string]any{"testField": 42},
			},
			wantValue: 42,
			wantErr:   false,
		},
		{
			name: "Valid integer string",
			n:    &Number{},
			args: args{
				ctx:   context.Background(),
				field: &mockField{name: "testField", fieldType: &Number{}},
				row:   map[string]any{"testField": "42"},
			},
			wantValue: 42,
			wantErr:   false,
		},
		{
			name: "Valid int pointer",
			n:    &Number{},
			args: args{
				ctx:   context.Background(),
				field: &mockField{name: "testField", fieldType: &Number{}},
				row:   map[string]any{"testField": &num},
			},
			wantValue: num,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Number{}
			gotValue, err := n.Scan(tt.args.ctx, tt.args.field, tt.args.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("Number.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("Number.Scan() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}

}
