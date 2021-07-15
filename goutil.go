package goutil

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/rc4"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
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
