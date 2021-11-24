package goutil

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/rc4"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Error struct {
	Code    int
	Message string
	Tm      time.Time
}

func (m *Error) Error() string {
	return m.Message
}

type ErrorHolder interface {
	GetError() error
	SetError(err error)
}

type DefaultErrorHolder struct {
	err error
}

func (m *DefaultErrorHolder) GetError() error {
	return m.err
}

func (m *DefaultErrorHolder) SetError(err error) {
	m.err = err
}

type Context interface {
	GetCtxData(key string) interface{}
	SetCtxData(key string, value interface{})
	RemoveCtxData(key string)
}

type DefaultContext struct {
	ctx     map[string]interface{}
	ctxLock sync.RWMutex
}

func (m *DefaultContext) GetCtxData(key string) interface{} {
	if key == "" {
		return nil
	}
	m.ctxLock.RLock()
	defer m.ctxLock.RUnlock()
	if m.ctx == nil {
		return nil
	}
	if ret, ok := m.ctx[key]; ok {
		return ret
	}
	return nil
}

func (m *DefaultContext) SetCtxData(key string, value interface{}) {
	if key == "" {
		return
	}
	m.ctxLock.Lock()
	defer m.ctxLock.Unlock()
	if m.ctx == nil {
		m.ctx = make(map[string]interface{})
	}
	m.ctx[key] = value
}

func (m *DefaultContext) RemoveCtxData(key string) {
	if key == "" {
		return
	}
	m.ctxLock.Lock()
	defer m.ctxLock.Unlock()
	if m.ctx == nil {
		return
	}
	delete(m.ctx, key)
}

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func FilePathExists(filePath string) bool {
	sts, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return sts.IsDir()
}

func FileModTime(file string) string {
	fi, err := os.Stat(file)
	if err != nil {
		return ""
	} else {
		return fi.ModTime().String()
	}
}

func SplitLR(s string, separator string) (l string, r string) {
	ret := strings.Split(s, separator)
	if len(ret) == 1 {
		return ret[0], ""
	} else {
		return ret[0], strings.Join(ret[1:], separator)
	}
}

func SplitByLine(s string) (ret []string) {
	if s == "" {
		return make([]string, 0)
	}
	ret3 := make([]string, 0)
	ret2 := make([]string, 0)
	var tmp []string
	ret1 := strings.Split(s, "\r\n")
	var j int
	j = len(ret1)
	for i := 0; i < j; i++ {
		tmp = strings.Split(ret1[i], "\r")
		x := len(tmp)
		for k := 0; k < x; k++ {
			ret2 = append(ret2, tmp[k])
		}
	}
	j = len(ret2)
	for i := 0; i < j; i++ {
		tmp = strings.Split(ret2[i], "\n")
		x := len(tmp)
		for k := 0; k < x; k++ {
			ret3 = append(ret3, tmp[k])
		}
	}
	return ret3
}

func FileMd5(fn string) (string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func BytesMd5(b []byte) (string, error) {
	h := md5.New()
	retn := 0
	for {
		btmp := b[retn:]
		n, e := h.Write(btmp)
		retn += n
		if e != nil && retn < len(b) {
			return "", e
		}
		if retn >= len(b) {
			break
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func StringMd5(s string) (string, error) {
	return BytesMd5([]byte(s))
}

func Md5(b []byte) ([]byte, error) {
	h := md5.New()
	retn := 0
	for {
		btmp := b[retn:]
		n, e := h.Write(btmp)
		retn += n
		if e != nil && retn < len(b) {
			return nil, e
		}
		if retn >= len(b) {
			break
		}
	}
	return h.Sum(nil), nil
}
func Gzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(data)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	zw.Flush()
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return buf.Bytes(), nil
}

func Gunzip(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	zr, eNew := gzip.NewReader(buf)
	if eNew != nil {
		return nil, eNew
	}
	var bufRet bytes.Buffer
	_, err := io.Copy(&bufRet, zr)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	zr.Close()
	return bufRet.Bytes(), nil
}

func Rc4(key []byte, data []byte) []byte {
	c, e := rc4.NewCipher(key)
	if e != nil {
		fmt.Println("fail.")
		return nil
	}
	bs := make([]byte, len(data))
	c.XORKeyStream(bs, data)
	return bs
}

// Deprecated: use syscall is Deprecated
func Daemonize(nochdir, noclose int) int {
	var ret, ret2 uintptr
	var err syscall.Errno

	darwin := runtime.GOOS == "darwin"

	// already a daemon
	if syscall.Getppid() == 1 {
		return 0
	}

	// fork off the parent process
	ret, ret2, err = syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	if err != 0 {
		return -1
	}

	// failure
	// if ret2 < 0 {
	// 	os.Exit(-1)
	// }

	// handle exception for darwin
	if darwin && ret2 == 1 {
		ret = 0
	}

	// if we got a good PID, then we call exit the parent process.
	if ret > 0 {
		os.Exit(0)
	}

	/* Change the file mode mask */
	_ = syscall.Umask(0)

	// create a new SID for the child process
	s_ret, s_errno := syscall.Setsid()
	if s_errno != nil {
		log.Printf("Error: syscall.Setsid errno: %d", s_errno)
	}
	if s_ret < 0 {
		return -1
	}

	if nochdir == 0 {
		os.Chdir("/")
	}

	if noclose == 0 {
		f, e := os.OpenFile("/dev/null", os.O_RDWR, 0)
		if e == nil {
			fd := f.Fd()
			syscall.Dup2(int(fd), int(os.Stdin.Fd()))
			syscall.Dup2(int(fd), int(os.Stdout.Fd()))
			syscall.Dup2(int(fd), int(os.Stderr.Fd()))
		}
	}

	return 0
}

func RestartProcess(keepFiles []*os.File, envVars []string, cleaner func() error) error {
	if cleaner != nil {
		if err := cleaner(); err != nil {
			return err
		}
	}

	// Use the original binary location. This works with symlinks such that if
	// the file it points to has been changed we will use the updated symlink.
	argv0, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}

	// working dir
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	//environment vars
	var env []string
	env = append(env, os.Environ()...)
	env = append(env, envVars...)

	allFiles := []*os.File{os.Stdin, os.Stdout, os.Stderr}
	allFiles = append(allFiles, keepFiles...)

	_, err = os.StartProcess(argv0, os.Args, &os.ProcAttr{
		Dir:   wd,
		Env:   env,
		Files: allFiles,
	})
	return err
}

func IsUtf8(bts []byte) bool {
	return utf8.Valid(bts)
}
func GBK2UTF8(srcGbk []byte) ([]byte, error) {
	gbkReader := bytes.NewReader(srcGbk)
	utf8Reader := transform.NewReader(gbkReader, simplifiedchinese.GBK.NewDecoder())
	return ioutil.ReadAll(utf8Reader)
}

func UTF82GBK(srcUtf8 []byte) ([]byte, error) {
	utf8Reader := bytes.NewReader(srcUtf8)
	gbkReader := transform.NewReader(utf8Reader, simplifiedchinese.GBK.NewEncoder())
	return ioutil.ReadAll(gbkReader)
}

func GetStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	rv := reflect.ValueOf(v)
	switch kd := rv.Kind(); kd {
	case reflect.Slice:
		return ""
	case reflect.String:
		return v.(string)
	case reflect.Float64, reflect.Float32:
		return fmt.Sprintf("%f", v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return fmt.Sprintf("%d", v)
	default:
		return ""
	}
}

func GetIntValueDefault(v interface{}, dft int64) int64 {
	if ret, ok := GetIntValue(v); !ok {
		return dft
	} else {
		return ret
	}
}

func GetIntValue(v interface{}) (int64, bool) {
	if v == nil {
		return 0, true
	}
	rv := reflect.ValueOf(v)
	switch kd := rv.Kind(); kd {
	case reflect.Slice:
		return 0, false
	case reflect.String:
		vs := v.(string)
		if ret, err := strconv.ParseInt(vs, 0, 64); err != nil {
			if retFloat, err := strconv.ParseFloat(vs, 64); err != nil {
				return 0, false
			} else {
				return int64(retFloat), true
			}
		} else {
			return ret, true
		}
	case reflect.Float64:
		return int64(v.(float64)), true
	case reflect.Float32:
		return int64(v.(float32)), true
	case reflect.Int:
		return int64(v.(int)), true
	case reflect.Int8:
		return int64(v.(int8)), true
	case reflect.Int16:
		return int64(v.(int16)), true
	case reflect.Int32:
		return int64(v.(int32)), true
	case reflect.Int64:
		return v.(int64), true
	case reflect.Uint:
		return int64(v.(uint)), true
	case reflect.Uint8:
		return int64(v.(uint8)), true
	case reflect.Uint16:
		return int64(v.(uint16)), true
	case reflect.Uint32:
		return int64(v.(uint32)), true
	case reflect.Uint64:
		return int64(v.(uint64)), true
	case reflect.Uintptr:
		return int64(v.(uintptr)), true
	default:
		return 0, false
	}
}

func GetFloatValue(v interface{}) (float64, bool) {
	if v == nil {
		return 0, true
	}
	rv := reflect.ValueOf(v)
	switch kd := rv.Kind(); kd {
	case reflect.Slice:
		return 0, false
	case reflect.String:
		vs := v.(string)
		if ret, err := strconv.ParseFloat(vs, 64); err != nil {
			return 0, false
		} else {
			return ret, true
		}
	case reflect.Float64:
		return v.(float64), true
	case reflect.Float32:
		return float64(v.(float32)), true
	case reflect.Int:
		return float64(v.(int)), true
	case reflect.Int8:
		return float64(v.(int8)), true
	case reflect.Int16:
		return float64(v.(int16)), true
	case reflect.Int32:
		return float64(v.(int32)), true
	case reflect.Int64:
		return float64(v.(int64)), true
	case reflect.Uint:
		return float64(v.(uint)), true
	case reflect.Uint8:
		return float64(v.(uint8)), true
	case reflect.Uint16:
		return float64(v.(uint16)), true
	case reflect.Uint32:
		return float64(v.(uint32)), true
	case reflect.Uint64:
		return float64(v.(uint64)), true
	case reflect.Uintptr:
		return float64(v.(uintptr)), true
	default:
		return 0, false
	}
}

func UrlEncodedMarshal(structObj interface{}) string {
	vl := reflect.ValueOf(structObj)
	if vl.Kind() == reflect.Ptr {
		vl = reflect.Indirect(vl)
	}
	if vl.Kind() != reflect.Struct {
		return ""
	}
	tp := vl.Type()
	ret := ""
	for i := 0; i < tp.NumField(); i++ {
		kv := "%s=%s"
		tag := tp.Field(i).Tag.Get("json")
		if tag == "" {
			tag = tp.Field(i).Name
		}
		value := GetStringValue(vl.Field(i).Interface())
		if value == "" {
			continue
		}
		tagList := strings.Split(tag, ",")
		if ret != "" {
			ret += "&"
		}
		ret += fmt.Sprintf(kv, tagList[0], url.QueryEscape(value))
	}
	return ret
}

type JsonInt int
type JsonStr string
type JsonFloat64 float64

func (m *JsonInt) UnmarshalJSON(bts []byte) error {
	if len(bts) == 0 {
		return nil
	}
	b := 0
	e := len(bts)
	if e > 2 {
		if bts[0] == '"' {
			b++
		}
		if bts[e-1] == '"' {
			e--
		}
	}
	if e == b {
		e++
	}
	i, err := strconv.Atoi(string(bts[b:e]))
	if err != nil {
		*m = 0
	} else {
		*m = JsonInt(i)
	}
	return nil
}

func (m *JsonStr) UnmarshalJSON(bts []byte) error {
	if len(bts) == 0 {
		return nil
	}
	b := 0
	e := len(bts)
	if e > 2 {
		if bts[0] == '"' {
			b++
		}
		if bts[e-1] == '"' {
			e--
		}
	}
	if e == b {
		e++
	}
	s := string(bts[b:e])
	*m = JsonStr(s)
	return nil
}

func (m *JsonFloat64) UnmarshalJSON(bts []byte) error {
	if len(bts) == 0 {
		return nil
	}
	b := 0
	e := len(bts)
	if e > 2 {
		if bts[0] == '"' {
			b++
		}
		if bts[e-1] == '"' {
			e--
		}
	}
	if e == b {
		e++
	}
	f, err := strconv.ParseFloat(string(bts[b:e]), 64)
	if err != nil {
		*m = 0
	} else {
		*m = JsonFloat64(f)
	}
	return nil
}

func (m *JsonInt) Scan(value interface{}) error {
	*m = 0
	if value == nil {
		return nil
	}
	switch s := value.(type) {
	case []uint8:
		ss := string(s)
		ii, err := strconv.ParseInt(ss, 10, 64)
		if err == nil {
			*m = JsonInt(ii)
		}
		return nil
	}
	r, b := GetIntValue(value)
	if b {
		*m = JsonInt(r)
	} else {
		*m = 0
	}
	return nil
}

func (m JsonInt) Value() (driver.Value, error) {
	return float64(m), nil
}

func (m *JsonStr) Scan(value interface{}) error {
	*m = ""
	if value == nil {
		return nil
	}
	switch s := value.(type) {
	case string:
		*m = JsonStr(s)
	case []uint8:
		*m = JsonStr(string(s))
	}
	return nil
}

func (m JsonStr) Value() (driver.Value, error) {
	return string(m), nil
}

func (m *JsonFloat64) Scan(value interface{}) error {
	*m = 0
	if value == nil {
		return nil
	}
	switch s := value.(type) {
	case []uint8:
		ss := string(s)
		ii, err := strconv.ParseFloat(ss, 64)
		if err == nil {
			*m = JsonFloat64(ii)
		}
		return nil
	}
	r, b := GetFloatValue(value)
	if b {
		*m = JsonFloat64(r)
	} else {
		*m = 0
	}
	return nil
}

func (m JsonFloat64) Value() (driver.Value, error) {
	return float64(m), nil
}

func LogStack(fd *os.File) {
	stack := make([]byte, 1024*1024*100)
	n := runtime.Stack(stack, true)
	if n > 0 {
		fd.WriteString(fmt.Sprintf("%s, numgoroutine: %d, process terminated, stack :\n", time.Now().Format("2006-01-02 15:04:05"), runtime.NumGoroutine()))
		fd.Write(stack[:n])
	} else {
		fd.WriteString(fmt.Sprintf("%s, numgoroutine: %d, get stack fail.", time.Now().Format("2006-01-02 15:04:05"), runtime.NumGoroutine()))
	}
}

func JsonMarshalNoHtmlEncode(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func InStringList(list []string, s string) bool {
	if len(list) == 0 {
		return false
	}
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}

func GetExePath() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	exePath, _ = filepath.EvalSymlinks(exePath)
	return filepath.Dir(exePath)
}

func GetExeName() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	exePath, _ = filepath.EvalSymlinks(exePath)
	return filepath.Base(exePath)
}
