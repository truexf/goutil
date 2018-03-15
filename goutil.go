package goutil

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
	"compress/gzip"
	"bytes"
	"crypto/rc4"
)

func leftRuneStr(s string, count int) string {
	cnt := 0
	for i, _ := range s {
		cnt++
		if cnt > count {
			fmt.Println(cnt)
			return s[:i]
		}
	}
	return s
}

func lenRuneStr(s string) int {
	ret := 0
	for i, _ := range s {
		ret++
		if i == 0 {
		}
	}
	return ret
}

func subRuneStr(s string, start int, count int) string {
	if s == "" || start < 0 || count <= 0 {
		return ""
	}
	cnt := 0
	idx := -1
	istart := -1
	iend := -1
	for i, _ := range s {
		idx++
		if istart == -1 && idx == start {
			istart = i
		}
		if istart > -1 {
			cnt++
		}
		if cnt > count {
			iend = i
			return s[istart:iend]
		}
	}
	if istart > -1 {
		if iend > -1 {
			return s[istart:iend]
		} else {
			return s[istart:]
		}
	}
	return ""
}

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
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
	tmp := make([]string, 0)
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
func Gzip(data []byte) ([]byte,error) {
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
	zr,eNew := gzip.NewReader(buf)
	if eNew != nil {
		return nil, eNew
	}
	var bufRet bytes.Buffer
	_, err := io.Copy(&bufRet,zr)
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
