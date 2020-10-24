package dagger

import (
	"encoding/json"
	"reflect"
)

// Values is a functional hash table for storing arbitrary data. It is not concurrency safe
type Values map[string]interface{}

// Exists returns true if the key exists in the Values
func (m Values) Exists(key string) bool {
	if _, ok := m[key]; ok {
		return true
	}
	return false
}

// Set set an entry in the Values
func (m Values) Set(k string, v interface{}) {
	if m == nil {
		m = map[string]interface{}{}
	}
	m[k] = v
}

// Get gets an entry from the Values by key
func (m Values) Get(key string) (interface{}, bool) {
	if m == nil {
		m = map[string]interface{}{}
	}
	if !m.Exists(key) {
		return nil, false
	}
	return m[key], true
}

// GetString gets an entry from the Values by key
func (m Values) GetString(key string) (string, bool) {
	if m == nil {
		m = map[string]interface{}{}
	}
	if !m.Exists(key) {
		return "", false
	}
	return m[key].(string), true
}

// Del deletes the entry from the Values by key
func (m Values) Del(key string) {
	if m == nil {
		m = map[string]interface{}{}
	}
	delete(m, key)
}

// Range iterates over the Values with the function. If the function returns false, the iteration exits.
func (m Values) Range(iterator func(key string, v interface{}) bool) {
	if m == nil {
		m = map[string]interface{}{}
	}
	for k, v := range m {
		if !iterator(k, v) {
			break
		}
	}
}

// Filter returns a Values of the values that return true from the filter function
func (m Values) Filter(filter func(key string, v interface{}) bool) Values {
	if m == nil {
		m = map[string]interface{}{}
	}
	data := Values{}
	if m == nil {
		return data
	}
	m.Range(func(key string, v interface{}) bool {
		if filter(key, v) {
			data.Set(key, v)
		}
		return true
	})
	return data
}

// Intersection returns the values that exist in both Valuess ref: https://en.wikipedia.org/wiki/Intersection_(set_theory)#:~:text=In%20mathematics%2C%20the%20intersection%20of,that%20also%20belong%20to%20A).
func (m Values) Intersection(other Values) Values {
	if m == nil {
		m = map[string]interface{}{}
	}
	toReturn := Values{}
	m.Range(func(key string, v interface{}) bool {
		if other.Exists(key) {
			toReturn.Set(key, v)
		}
		return true
	})
	return toReturn
}

// Union returns the all values in both Valuess ref: https://en.wikipedia.org/wiki/Union_(set_theory)
func (m Values) Union(other Values) Values {
	toReturn := Values{}
	m.Range(func(k string, v interface{}) bool {
		toReturn.Set(k, v)
		return true
	})
	other.Range(func(k string, v interface{}) bool {
		toReturn.Set(k, v)
		return true
	})
	return toReturn
}

// Copy creates a replica of the Values
func (m Values) Copy() Values {
	copied := Values{}
	if m == nil {
		return copied
	}
	m.Range(func(k string, v interface{}) bool {
		copied.Set(k, v)
		return true
	})
	return copied
}

func (v Values) Equals(other Values) bool {
	return reflect.DeepEqual(v, other)
}

func (v Values) GetNested(key string) (Values, bool) {
	if val, ok := v[key]; ok && val != nil {
		if values, ok := val.(Values); ok {
			return values, true
		}
	}
	return nil, false
}

func (v Values) IsNested(key string) bool {
	_, ok := v.GetNested(key)
	return ok
}

func (v Values) SetNested(key string, values Values) {
	v.Set(key, values)
}

func (v Values) JSON() ([]byte, error) {
	return json.Marshal(v)
}

func (m Values) FromJSON(data []byte) error {
	return json.Unmarshal(data, &m)
}

func (m Values) UnmarshalFrom(obj json.Marshaler) error {
	bits, err := obj.MarshalJSON()
	if err != nil {
		return err
	}
	return m.FromJSON(bits)
}

func (m Values) MarshalTo(obj json.Unmarshaler) error {
	bits, err := m.JSON()
	if err != nil {
		return err
	}
	return obj.UnmarshalJSON(bits)
}
