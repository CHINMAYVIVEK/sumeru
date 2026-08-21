package modelmeta

import "reflect"

// ModelNameFromStruct reads the technical model name from an embedded ModelMeta tag,
// or falls back to ModelNameFromGo for the struct name.
func ModelNameFromStruct(st reflect.Type) (string, error) {
	for st.Kind() == reflect.Pointer {
		st = st.Elem()
	}
	if st.Kind() != reflect.Struct {
		return "", nil
	}
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		if !f.Anonymous {
			continue
		}
		t := f.Type
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t != reflect.TypeOf(ModelMeta{}) {
			continue
		}
		raw, err := ParseModelTag(string(f.Tag.Get("sumeru")))
		if err != nil {
			return "", err
		}
		if raw == "-" {
			return "-", nil
		}
		if raw != "" {
			return raw, nil
		}
		break
	}
	return ModelNameFromGo(st.Name()), nil
}
