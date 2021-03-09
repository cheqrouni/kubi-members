package utils

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"reflect"
	"strings"
)

const tagName = "ldap"

func Unmarshal(entry *ldap.Entry, i interface{}) (err error) {

	// Make sure it's a ptr
	if reflect.ValueOf(i).Kind() != reflect.Ptr {
		return fmt.Errorf("expecting a ptr not %s", reflect.ValueOf(i).Kind())

	}
	// Get ptr value and type
	v, t := reflect.ValueOf(i).Elem(), reflect.TypeOf(i).Elem()

	// Make sure it's pointing to a struct
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expecting a ptr to a struct not %s", v.Kind())
	}

	// Go trough all the struct fields
	for n := 0; n < t.NumField(); n++ {
		// Holds field value and type
		fv, ft := v.Field(n), t.Field(n)
		// skip unexported fields
		if ft.PkgPath != "" {
			continue
		}

		// Read tag from field type
		tag, _ := readTag(ft)

		// All entry have dn
		if tag == "dn" {
			fv.SetString(entry.DN)
			continue
		}

		// Go trough other attributes
		for _, attr := range entry.Attributes {

			// Match attributes
			if attr.Name == tag {

				// Result is empty
				if len(attr.Values) == 0 {
					continue
				}

				// Result is a string
				if len(attr.Values) == 1 && fv.Kind() == reflect.String {
					fv.SetString(attr.Values[0])
					continue
				}

				// Result is a slice
				if len(attr.Values) > 1 && fv.Kind() == reflect.Slice {
					for _, item := range attr.Values {
						fv.Set(reflect.Append(fv, reflect.ValueOf(item)))
					}
				}

				// TODO : Add nested struct ?
				//if len(attr.Values) > 1 && fv.Kind() == reflect.Ptr && reflect.TypeOf(fv).Elem().Kind() == reflect.Struct {
				//
				//}

			}
		}
	}
	return
}

func readTag(f reflect.StructField) (string, bool) {
	val, ok := f.Tag.Lookup(tagName)
	if !ok {
		return f.Name, false
	}
	opts := strings.Split(val, ",")
	omit := false
	if len(opts) == 2 {
		omit = opts[1] == "omitempty"
	}
	return opts[0], omit
}
