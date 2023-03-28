package mysqlutil

import (
	"database/sql/driver"
	"github.com/truexf/goutil"
	"strconv"
)

type NullInt int64
type NullString string
type NullFloat float64

func (m *NullInt) Scan(value interface{}) error {
	if value == nil {
		*m = 0
		return nil
	}
	switch s := value.(type) {
	case []uint8:
		ss := string(s)
		ii, err := strconv.ParseInt(ss, 10, 64)
		if err == nil {
			*m = NullInt(ii)
		}
		return nil
	}
	r, b := goutil.GetIntValue(value)
	if b {
		*m = NullInt(r)
	} else {
		*m = 0
	}
	return nil
}

func (m NullInt) Value() (driver.Value, error) {
	return float64(m), nil
}

func (m *NullString) Scan(value interface{}) error {
	if value == nil {
		*m = ""
		return nil
	}
	switch s := value.(type) {
	case string:
		*m = NullString(s)
	case []uint8:
		*m = NullString(string(s))
	}
	return nil
}

func (m NullString) Value() (driver.Value, error) {
	return string(m), nil
}

func (m *NullFloat) Scan(value interface{}) error {
	if value == nil {
		*m = 0
		return nil
	}
	switch s := value.(type) {
	case []uint8:
		ss := string(s)
		ii, err := strconv.ParseFloat(ss, 64)
		if err == nil {
			*m = NullFloat(ii)
		}
		return nil
	}
	r, b := goutil.GetFloatValue(value)
	if b {
		*m = NullFloat(r)
	} else {
		*m = 0
	}
	return nil
}

func (m NullFloat) Value() (driver.Value, error) {
	return float64(m), nil
}
