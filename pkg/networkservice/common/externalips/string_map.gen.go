// Code generated by "go-syncmap -output string_map.gen.go -type stringMap<string,string>"; DO NOT EDIT.

package externalips

import "sync"

func _() {
	// An "cannot convert stringMap literal (type stringMap) to type sync.Map" compiler error signifies that the base type have changed.
	// Re-run the go-syncmap command to generate them again.
	_ = (sync.Map)(stringMap{})
}

var _nil_stringMap_string_value = func() (val string) { return }()

func (m *stringMap) Store(key string, value string) {
	(*sync.Map)(m).Store(key, value)
}

func (m *stringMap) LoadOrStore(key string, value string) (string, bool) {
	actual, loaded := (*sync.Map)(m).LoadOrStore(key, value)
	if actual == nil {
		return _nil_stringMap_string_value, loaded
	}
	return actual.(string), loaded
}

func (m *stringMap) Load(key string) (string, bool) {
	value, ok := (*sync.Map)(m).Load(key)
	if value == nil {
		return _nil_stringMap_string_value, ok
	}
	return value.(string), ok
}

func (m *stringMap) Delete(key string) {
	(*sync.Map)(m).Delete(key)
}

func (m *stringMap) Range(f func(key string, value string) bool) {
	(*sync.Map)(m).Range(func(key, value interface{}) bool {
		return f(key.(string), value.(string))
	})
}
