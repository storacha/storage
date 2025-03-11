package enum

// inspiration: https://github.com/zarldev/goenums

import "fmt"

type KeyFormat struct {
	keyFormat
}

type keyFormat int

const (
	unknown keyFormat = iota
	json
	pem
)

var (
	strOperationMap = map[keyFormat]string{
		json: "JSON",
		pem:  "PEM",
	}

	typeOperationMap = map[string]keyFormat{
		"JSON": json,
		"PEM":  pem,
	}
)

func (t keyFormat) String() string {
	return strOperationMap[t]
}

func ParseKeyFormat(a any) KeyFormat {
	switch v := a.(type) {
	case KeyFormat:
		return v
	case string:
		return KeyFormat{stringToOperation(v)}
	case fmt.Stringer:
		return KeyFormat{stringToOperation(v.String())}
	case int:
		return KeyFormat{keyFormat(v)}
	case int64:
		return KeyFormat{keyFormat(int(v))}
	case int32:
		return KeyFormat{keyFormat(int(v))}
	}
	return KeyFormat{unknown}
}

func stringToOperation(s string) keyFormat {
	if v, ok := typeOperationMap[s]; ok {
		return v
	}
	return unknown
}

func (t keyFormat) IsValid() bool {
	_, ok := strOperationMap[t]
	return ok
}

type keyFormatContainer struct {
	UNKNOWN KeyFormat
	JSON    KeyFormat
	PEM     KeyFormat
}

var KeyFormats = keyFormatContainer{
	UNKNOWN: KeyFormat{unknown},
	JSON:    KeyFormat{json},
	PEM:     KeyFormat{pem},
}

func (c keyFormatContainer) All() []KeyFormat {
	return []KeyFormat{
		c.JSON,
		c.PEM,
	}
}

func (t KeyFormat) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *KeyFormat) UnmarshalJSON(b []byte) error {
	*t = ParseKeyFormat(string(b))
	return nil
}
