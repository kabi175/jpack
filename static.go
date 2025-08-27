package jpack

import (
	"errors"
	"reflect"
	"strings"

	"github.com/samber/lo"
)

type FieldTypeResolver func(field reflect.StructField) (JFieldType, error)

var DefaultSchemaRegistry = map[string]JSchema{}

var DefaultFieldTypeRegistry = map[reflect.Kind]func(field reflect.StructField) (JFieldType, error){
	reflect.String:  func(field reflect.StructField) (JFieldType, error) { return &String{}, nil },
	reflect.Int:     func(field reflect.StructField) (JFieldType, error) { return &Number{}, nil },
	reflect.Float64: func(field reflect.StructField) (JFieldType, error) { return &Number{}, nil },
	reflect.Bool:    func(field reflect.StructField) (JFieldType, error) { return &Boolean{}, nil },
	reflect.Struct: func(field reflect.StructField) (JFieldType, error) {

		return &Ref{}, nil
	},
}

var CustomFieldTypeRegistry = map[string]func(field reflect.StructField) (JFieldType, error){
	"ref":     func(field reflect.StructField) (JFieldType, error) { return &Ref{}, nil },
	"string":  func(field reflect.StructField) (JFieldType, error) { return &String{}, nil },
	"number":  func(field reflect.StructField) (JFieldType, error) { return &Number{}, nil },
	"boolean": func(field reflect.StructField) (JFieldType, error) { return &Boolean{}, nil },
}

func DefaultFieldTypeResolver(field reflect.StructField) (JFieldType, error) {
	tags := strings.Split(field.Tag.Get("jpack"), ",")
	typeTag, ok := lo.Find(tags, func(tag string) bool {
		return strings.HasPrefix(tag, "type:")
	})

	if ok {
		resolver, ok := CustomFieldTypeRegistry[typeTag]

		if !ok {
			return nil, errors.New("invalid field type")
		}

		return resolver(field)
	}

	fkind := field.Type
	if fkind.Kind() == reflect.Ptr {
		fkind = field.Type.Elem()
	}

	resolver, ok := DefaultFieldTypeRegistry[fkind.Kind()]

	if !ok {
		return nil, errors.New("invalid field type")
	}

	return resolver(field)
}

func DefaultFieldResolver(schema JSchema, field reflect.StructField) error {
	fieldType, err := DefaultFieldTypeResolver(field)

	if err != nil {
		return err
	}

	fieldName := field.Name
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		fieldName = strings.Split(jsonTag, ",")[0]
	}

	if _, ok := fieldType.(*Ref); ok {
		tags := field.Tag.Get("jpack")

		ref, ok := lo.Find(strings.Split(tags, ","), func(tag string) bool {
			return strings.HasPrefix(tag, "ref:")
		})

		if !ok {
			return errors.New("invalid field type")
		}

		ref = strings.TrimPrefix(ref, "ref:")

		schema, ok := DefaultSchemaRegistry[ref]

		if !ok {
			return errors.New("schema not found")
		}

		schema.AddRef(fieldName, schema)
		return nil
	}

	schema.AddField(fieldName, fieldType, nil)

	return nil
}

func SchemaFor(name string, obj any) (JSchema, error) {
	if obj == nil {
		return nil, errors.New("object is nil")
	}

	schema := NewSchema(name).Build()
	// use reflection to get the fields of the object
	typ := reflect.TypeOf(obj)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		DefaultFieldResolver(schema, field)
	}

	return schema, nil

}

type StaticSchema struct {
	schema JSchema
}
