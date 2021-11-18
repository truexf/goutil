// Copyright 2021 fangyousong(方友松). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mysqlutil

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const digits01 = "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
const digits10 = "0000000000111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999"

func reserveBuffer(buf []byte, appendSize int) []byte {
	newSize := len(buf) + appendSize
	if cap(buf) < newSize {
		// Grow buffer exponentially
		newBuf := make([]byte, len(buf)*2+appendSize)
		copy(newBuf, buf)
		buf = newBuf
	}
	return buf[:newSize]
}

func escapeBytesBackslash(buf, v []byte) []byte {
	pos := len(buf)
	buf = reserveBuffer(buf, len(v)*2)

	for _, c := range v {
		switch c {
		case '\x00':
			buf[pos] = '\\'
			buf[pos+1] = '0'
			pos += 2
		case '\n':
			buf[pos] = '\\'
			buf[pos+1] = 'n'
			pos += 2
		case '\r':
			buf[pos] = '\\'
			buf[pos+1] = 'r'
			pos += 2
		case '\x1a':
			buf[pos] = '\\'
			buf[pos+1] = 'Z'
			pos += 2
		case '\'':
			buf[pos] = '\\'
			buf[pos+1] = '\''
			pos += 2
		case '"':
			buf[pos] = '\\'
			buf[pos+1] = '"'
			pos += 2
		case '\\':
			buf[pos] = '\\'
			buf[pos+1] = '\\'
			pos += 2
		default:
			buf[pos] = c
			pos++
		}
	}

	return buf[:pos]
}

func escapeBytesQuotes(buf, v []byte) []byte {
	pos := len(buf)
	buf = reserveBuffer(buf, len(v)*2)

	for _, c := range v {
		if c == '\'' {
			buf[pos] = '\''
			buf[pos+1] = '\''
			pos += 2
		} else {
			buf[pos] = c
			pos++
		}
	}

	return buf[:pos]
}

func escapeStringBackslash(buf []byte, v string) []byte {
	pos := len(buf)
	buf = reserveBuffer(buf, len(v)*2)

	for i := 0; i < len(v); i++ {
		c := v[i]
		switch c {
		case '\x00':
			buf[pos] = '\\'
			buf[pos+1] = '0'
			pos += 2
		case '\n':
			buf[pos] = '\\'
			buf[pos+1] = 'n'
			pos += 2
		case '\r':
			buf[pos] = '\\'
			buf[pos+1] = 'r'
			pos += 2
		case '\x1a':
			buf[pos] = '\\'
			buf[pos+1] = 'Z'
			pos += 2
		case '\'':
			buf[pos] = '\\'
			buf[pos+1] = '\''
			pos += 2
		case '"':
			buf[pos] = '\\'
			buf[pos+1] = '"'
			pos += 2
		case '\\':
			buf[pos] = '\\'
			buf[pos+1] = '\\'
			pos += 2
		default:
			buf[pos] = c
			pos++
		}
	}

	return buf[:pos]
}

func escapeStringQuotes(buf []byte, v string) []byte {
	pos := len(buf)
	buf = reserveBuffer(buf, len(v)*2)

	for i := 0; i < len(v); i++ {
		c := v[i]
		if c == '\'' {
			buf[pos] = '\''
			buf[pos+1] = '\''
			pos += 2
		} else {
			buf[pos] = c
			pos++
		}
	}

	return buf[:pos]
}

func MySqlEscape(noBackslashEscapes bool, v interface{}) (string, error) {
	buf := make([]byte, 0)
	buf = buf[:0]
	switch v := v.(type) {
	case int64:
		buf = strconv.AppendInt(buf, v, 10)
	case int:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case float64:
		buf = strconv.AppendFloat(buf, v, 'g', -1, 64)
	case float32:
		buf = strconv.AppendFloat(buf, float64(v), 'g', -1, 64)
	case bool:
		if v {
			buf = append(buf, '1')
		} else {
			buf = append(buf, '0')
		}
	case time.Time:
		if v.IsZero() {
			buf = append(buf, "'0000-00-00'"...)
		} else {
			v = v.Add(time.Nanosecond * 500) // To round under microsecond
			year := v.Year()
			year100 := year / 100
			year1 := year % 100
			month := v.Month()
			day := v.Day()
			hour := v.Hour()
			minute := v.Minute()
			second := v.Second()
			micro := v.Nanosecond() / 1000

			buf = append(buf, []byte{
				'\'',
				digits10[year100], digits01[year100],
				digits10[year1], digits01[year1],
				'-',
				digits10[month], digits01[month],
				'-',
				digits10[day], digits01[day],
				' ',
				digits10[hour], digits01[hour],
				':',
				digits10[minute], digits01[minute],
				':',
				digits10[second], digits01[second],
			}...)

			if micro != 0 {
				micro10000 := micro / 10000
				micro100 := micro / 100 % 100
				micro1 := micro % 100
				buf = append(buf, []byte{
					'.',
					digits10[micro10000], digits01[micro10000],
					digits10[micro100], digits01[micro100],
					digits10[micro1], digits01[micro1],
				}...)
			}
			buf = append(buf, '\'')
		}
	case []byte:
		if v == nil {
			buf = append(buf, "NULL"...)
		} else {
			buf = append(buf, "_binary'"...)
			if !noBackslashEscapes {
				buf = escapeBytesBackslash(buf, v)
			} else {
				buf = escapeBytesQuotes(buf, v)
			}
			buf = append(buf, '\'')
		}
	case string:
		buf = append(buf, '\'')
		if !noBackslashEscapes {
			buf = escapeStringBackslash(buf, v)
		} else {
			buf = escapeStringQuotes(buf, v)
		}
		buf = append(buf, '\'')
	default:
		vk := reflect.ValueOf(v).Kind()
		switch vk {
		case reflect.String:
			sValue := reflect.ValueOf(v).String()
			buf = append(buf, '\'')
			if !noBackslashEscapes {
				buf = escapeStringBackslash(buf, sValue)
			} else {
				buf = escapeStringQuotes(buf, sValue)
			}
			buf = append(buf, '\'')
		case reflect.Float64, reflect.Float32:
			fValue := reflect.ValueOf(v).Float()
			buf = strconv.AppendFloat(buf, fValue, 'g', -1, 64)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			iValue := reflect.ValueOf(v).Int()
			buf = strconv.AppendInt(buf, iValue, 10)
		default:
			return "", driver.ErrSkip
		}
	}
	return string(buf), nil
}

func MySqlEscapeDefault(noBackslashEscapes bool, v interface{}, defaultRet string) string {
	ret, err := MySqlEscape(noBackslashEscapes, v)
	if err != nil {
		return defaultRet
	}
	return ret
}

func MySqlMogrify(noBackslashEscapes bool, query string, args ...interface{}) (string, error) {
	if strings.Count(query, "?") != len(args) {
		return "", driver.ErrSkip
	}

	buf := make([]byte, 0)
	buf = buf[:0]
	argPos := 0

	for i := 0; i < len(query); i++ {
		q := strings.IndexByte(query[i:], '?')
		if q == -1 {
			buf = append(buf, query[i:]...)
			break
		}
		buf = append(buf, query[i:i+q]...)
		i += q

		arg := args[argPos]
		argPos++

		if arg == nil {
			buf = append(buf, "NULL"...)
			continue
		}

		switch v := arg.(type) {
		case int64:
			buf = strconv.AppendInt(buf, v, 10)
		case int:
			buf = strconv.AppendInt(buf, int64(v), 10)
		case float64:
			buf = strconv.AppendFloat(buf, v, 'g', -1, 64)
		case float32:
			buf = strconv.AppendFloat(buf, float64(v), 'g', -1, 64)
		case bool:
			if v {
				buf = append(buf, '1')
			} else {
				buf = append(buf, '0')
			}
		case time.Time:
			if v.IsZero() {
				buf = append(buf, "'0000-00-00'"...)
			} else {
				v = v.Add(time.Nanosecond * 500) // To round under microsecond
				year := v.Year()
				year100 := year / 100
				year1 := year % 100
				month := v.Month()
				day := v.Day()
				hour := v.Hour()
				minute := v.Minute()
				second := v.Second()
				micro := v.Nanosecond() / 1000

				buf = append(buf, []byte{
					'\'',
					digits10[year100], digits01[year100],
					digits10[year1], digits01[year1],
					'-',
					digits10[month], digits01[month],
					'-',
					digits10[day], digits01[day],
					' ',
					digits10[hour], digits01[hour],
					':',
					digits10[minute], digits01[minute],
					':',
					digits10[second], digits01[second],
				}...)

				if micro != 0 {
					micro10000 := micro / 10000
					micro100 := micro / 100 % 100
					micro1 := micro % 100
					buf = append(buf, []byte{
						'.',
						digits10[micro10000], digits01[micro10000],
						digits10[micro100], digits01[micro100],
						digits10[micro1], digits01[micro1],
					}...)
				}
				buf = append(buf, '\'')
			}
		case []byte:
			if v == nil {
				buf = append(buf, "NULL"...)
			} else {
				buf = append(buf, "_binary'"...)
				if !noBackslashEscapes {
					buf = escapeBytesBackslash(buf, v)
				} else {
					buf = escapeBytesQuotes(buf, v)
				}
				buf = append(buf, '\'')
			}
		case string:
			buf = append(buf, '\'')
			if !noBackslashEscapes {
				buf = escapeStringBackslash(buf, v)
			} else {
				buf = escapeStringQuotes(buf, v)
			}
			buf = append(buf, '\'')
		default:
			vk := reflect.ValueOf(arg).Kind()
			switch vk {
			case reflect.String:
				sValue := reflect.ValueOf(arg).String()
				buf = append(buf, '\'')
				if !noBackslashEscapes {
					buf = escapeStringBackslash(buf, sValue)
				} else {
					buf = escapeStringQuotes(buf, sValue)
				}
				buf = append(buf, '\'')
			case reflect.Float64, reflect.Float32:
				fValue := reflect.ValueOf(arg).Float()
				buf = strconv.AppendFloat(buf, fValue, 'g', -1, 64)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				iValue := reflect.ValueOf(arg).Int()
				buf = strconv.AppendInt(buf, iValue, 10)
			default:
				return "", driver.ErrSkip
			}
		}
	}
	if argPos != len(args) {
		return "", driver.ErrSkip
	}
	return string(buf), nil
}

func UpsertSql(dbTaggedObj interface{}, tableName string, keyFields []string) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("tableName is nil")
	}

	vl := reflect.ValueOf(dbTaggedObj)
	if vl.Kind() == reflect.Ptr {
		vl = reflect.Indirect(vl)
	}
	if vl.Kind() != reflect.Struct {
		return "", fmt.Errorf("dbTaggedObj is not a struct")
	}
	tp := vl.Type()
	fv := make(map[string]interface{})
	for i := 0; i < tp.NumField(); i++ {
		dbTag := tp.Field(i).Tag.Get("db")
		if dbTag == "" {
			continue
		}
		fv[dbTag] = vl.FieldByName(tp.Field(i).Name).Interface()
	}
	kfv := make(map[string]interface{})
	for _, k := range keyFields {
		if v, ok := fv[k]; ok {
			kfv[k] = v
		} else {
			return "", fmt.Errorf("key field [%s] not found", k)
		}
	}
	sqlFields := ""
	sqlParams := ""
	sqlValues := make([]interface{}, 0)
	sqlUpdate := ""
	sqlUpdateValues := make([]interface{}, 0)
	for k, v := range fv {
		if sqlFields != "" {
			sqlFields += ", "
			sqlParams += ", "
		}
		sqlFields += k
		sqlParams += "?"
		sqlValues = append(sqlValues, v)
		if _, ok := kfv[k]; !ok {
			if sqlUpdate != "" {
				sqlUpdate += ", "
			}
			sqlUpdate += fmt.Sprintf("%s = ?", k)
			sqlUpdateValues = append(sqlUpdateValues, v)
		}
	}
	sqlValues = append(sqlValues, sqlUpdateValues...)
	var sql string
	if sqlUpdate != "" {
		sql = fmt.Sprintf("insert into %s (%s) values (%s) on duplicate key update %s ", tableName, sqlFields, sqlParams, sqlUpdate)
	} else {
		sql = fmt.Sprintf("insert into %s (%s) values (%s)", tableName, sqlFields, sqlParams)
	}
	return MySqlMogrify(false, sql, sqlValues...)
}

func InsertSql(dbTaggedObj interface{}, tableName string) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("tableName is nil")
	}

	vl := reflect.ValueOf(dbTaggedObj)
	if vl.Kind() == reflect.Ptr {
		vl = reflect.Indirect(vl)
	}
	if vl.Kind() != reflect.Struct {
		return "", fmt.Errorf("dbTaggedObj is not a struct")
	}
	tp := vl.Type()
	fv := make(map[string]interface{})
	for i := 0; i < tp.NumField(); i++ {
		dbTag := tp.Field(i).Tag.Get("db")
		if dbTag == "" {
			continue
		}
		fv[dbTag] = vl.FieldByName(tp.Field(i).Name).Interface()
	}
	sqlFields := ""
	sqlParams := ""
	sqlValues := make([]interface{}, 0)
	for k, v := range fv {
		if sqlFields != "" {
			sqlFields += ", "
			sqlParams += ", "
		}
		sqlFields += k
		sqlParams += "?"
		sqlValues = append(sqlValues, v)
	}
	sql := fmt.Sprintf("insert into %s (%s) values (%s)", tableName, sqlFields, sqlParams)

	return MySqlMogrify(false, sql, sqlValues...)
}

func DeleteSql(dbTaggedObj interface{}, tableName string) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("tableName is nil")
	}

	vl := reflect.ValueOf(dbTaggedObj)
	if vl.Kind() == reflect.Ptr {
		vl = reflect.Indirect(vl)
	}
	if vl.Kind() != reflect.Struct {
		return "", fmt.Errorf("dbTaggedObj is not a struct")
	}
	tp := vl.Type()
	fv := make(map[string]interface{})
	for i := 0; i < tp.NumField(); i++ {
		dbTag := tp.Field(i).Tag.Get("db")
		if dbTag == "" {
			continue
		}
		fv[dbTag] = vl.FieldByName(tp.Field(i).Name).Interface()
	}
	sqlFields := ""
	sqlValues := make([]interface{}, 0)
	for k, v := range fv {
		if sqlFields != "" {
			sqlFields += " and "
		}
		sqlFields += fmt.Sprintf("%s = ?", k)
		sqlValues = append(sqlValues, v)
	}
	sql := fmt.Sprintf("delete from %s where %s", tableName, sqlFields)

	return MySqlMogrify(false, sql, sqlValues...)
}

func SelectSql(dbTaggedObj interface{}, tableName string, selectFields string) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("tableName is nil")
	}

	vl := reflect.ValueOf(dbTaggedObj)
	if vl.Kind() == reflect.Ptr {
		vl = reflect.Indirect(vl)
	}
	if vl.Kind() != reflect.Struct {
		return "", fmt.Errorf("dbTaggedObj is not a struct")
	}
	tp := vl.Type()
	fv := make(map[string]interface{})
	for i := 0; i < tp.NumField(); i++ {
		dbTag := tp.Field(i).Tag.Get("db")
		if dbTag == "" {
			continue
		}
		fv[dbTag] = vl.FieldByName(tp.Field(i).Name).Interface()
	}
	sqlFields := ""
	sqlValues := make([]interface{}, 0)
	for k, v := range fv {
		if sqlFields != "" {
			sqlFields += " and "
		}
		sqlFields += fmt.Sprintf("%s = ?", k)
		sqlValues = append(sqlValues, v)
	}
	if selectFields == "" {
		selectFields = "*"
	}
	sql := fmt.Sprintf("select %s from %s where %s", selectFields, tableName, sqlFields)

	return MySqlMogrify(false, sql, sqlValues...)
}

func UpsertSqlM(objMap map[string]interface{}, tableName string, keyFields []string) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("tableName is nil")
	}
	if objMap == nil {
		return "", fmt.Errorf("objMap is nil")
	}

	keyFieldValue := make(map[string]interface{})
	for _, k := range keyFields {
		if v, ok := objMap[k]; ok {
			keyFieldValue[k] = v
		} else {
			return "", fmt.Errorf("key field [%s] not found", k)
		}
	}
	sqlFields := ""
	sqlParams := ""
	sqlValues := make([]interface{}, 0)
	sqlUpdate := ""
	sqlUpdateValues := make([]interface{}, 0)
	for k, v := range objMap {
		if sqlFields != "" {
			sqlFields += ", "
			sqlParams += ", "
		}
		sqlFields += k
		sqlParams += "?"
		sqlValues = append(sqlValues, v)
		if _, ok := keyFieldValue[k]; !ok {
			if sqlUpdate != "" {
				sqlUpdate += ", "
			}
			sqlUpdate += fmt.Sprintf("%s = ?", k)
			sqlUpdateValues = append(sqlUpdateValues, v)
		}
	}
	sqlValues = append(sqlValues, sqlUpdateValues...)
	var sql string
	if sqlUpdate != "" {
		sql = fmt.Sprintf("insert into %s (%s) values (%s) on duplicate key update %s ", tableName, sqlFields, sqlParams, sqlUpdate)
	} else {
		sql = fmt.Sprintf("insert into %s (%s) values (%s)", tableName, sqlFields, sqlParams)
	}
	return MySqlMogrify(false, sql, sqlValues...)
}

func InsertSqlM(objMap map[string]interface{}, tableName string) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("tableName is nil")
	}
	if objMap == nil {
		return "", fmt.Errorf("objMap is nil")
	}

	sqlFields := ""
	sqlParams := ""
	sqlValues := make([]interface{}, 0)
	for k, v := range objMap {
		if sqlFields != "" {
			sqlFields += ", "
			sqlParams += ", "
		}
		sqlFields += k
		sqlParams += "?"
		sqlValues = append(sqlValues, v)
	}
	sql := fmt.Sprintf("insert into %s (%s) values (%s)", tableName, sqlFields, sqlParams)
	return MySqlMogrify(false, sql, sqlValues...)
}

func DeleteSqlM(objMap map[string]interface{}, tableName string) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("tableName is nil")
	}
	if objMap == nil {
		return "", fmt.Errorf("objMap is nil")
	}

	sqlFields := ""
	sqlValues := make([]interface{}, 0)
	for k, v := range objMap {
		if sqlFields != "" {
			sqlFields += " and "
		}
		sqlFields += fmt.Sprintf("%s = ?", k)
		sqlValues = append(sqlValues, v)
	}
	sql := fmt.Sprintf("delete from %s where %s", tableName, sqlFields)
	return MySqlMogrify(false, sql, sqlValues...)
}

func SelectSqlM(objMap map[string]interface{}, tableName string, selectFields string) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("tableName is nil")
	}
	if objMap == nil {
		return "", fmt.Errorf("objMap is nil")
	}

	sqlFields := ""
	sqlValues := make([]interface{}, 0)
	for k, v := range objMap {
		if sqlFields != "" {
			sqlFields += " and "
		}
		sqlFields += fmt.Sprintf("%s = ?", k)
		sqlValues = append(sqlValues, v)
	}
	if selectFields == "" {
		selectFields = "*"
	}
	sql := fmt.Sprintf("select %s from %s where %s", selectFields, tableName, sqlFields)
	return MySqlMogrify(false, sql, sqlValues...)
}
