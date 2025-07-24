package jpack

import "github.com/samber/lo"

func PK(schema JSchema) (JField, bool) {
	return lo.Find(schema.Fields(), func(f JField) bool {
		return f.Name() == "id"
	})
}
