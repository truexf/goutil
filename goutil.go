package goutil

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
)

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
