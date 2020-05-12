package common

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type BufferedFileWriter struct {
	incomingQueue      *LinkedList
	incomingNotifyChan chan int
	flushChan          chan *bytes.Buffer
	freeChan           chan int
	buffer1            bytes.Buffer
	buffer2            bytes.Buffer
	bufferCurrent      *bytes.Buffer
	bufferCap          int
	flushInterval      time.Duration
	fileTimeMask       string
	filePath           string
	terminalSignal     chan int
	stopped            bool
	terminalDone       chan int
	flushSignal        chan int
	flushDone          chan int
}

//fileTimeMask: use golang time-format
func NewBufferedFileWriter(bufferSize int, flushIntervalSecond int, filePath string, fileTimeMask string) (*BufferedFileWriter, error) {
	ret := &BufferedFileWriter{}
	if bufferSize < 1024 || bufferSize > 1024*1024*1024 {
		return nil, fmt.Errorf("1kbytes < buffer-size < 1gbytes")
	}

	if flushIntervalSecond <= 0 {
		ret.flushInterval = time.Second * 10
	} else {
		ret.flushInterval = time.Second * time.Duration(flushIntervalSecond)
	}

	if filePath == "" {
		return nil, fmt.Errorf("file-path is empty")
	}
	stat, err := os.Stat(filePath)
	if err != nil || !stat.IsDir() {
		return nil, fmt.Errorf("invalid file-path: %s", filePath)
	}
	ret.filePath = filePath

	if fileTimeMask == "" {
		return nil, fmt.Errorf("file-time-mask is empty")
	}
	ret.fileTimeMask = fileTimeMask

	ret.buffer1.Grow(bufferSize)
	ret.buffer2.Grow(bufferSize)
	ret.bufferCap = bufferSize
	ret.bufferCurrent = &ret.buffer1
	ret.incomingQueue = NewLinkedList(true)
	ret.incomingNotifyChan = make(chan int, 10000)
	ret.flushChan = make(chan *bytes.Buffer)
	ret.freeChan = make(chan int, 2)
	ret.freeChan <- 1
	ret.terminalSignal = make(chan int)
	ret.terminalDone = make(chan int)
	ret.flushSignal = make(chan int)
	ret.flushDone = make(chan int)

	go func() {
		for {
			ret.flush()
		}
	}()

	go ret.run()

	go func() {
		<- time.After(ret.flushInterval)
		ret.Flush(0)
	}()

	return ret, nil
}

func (m *BufferedFileWriter) Write(bts []byte) error {
	if m.stopped {
		return fmt.Errorf("i was terminated")
	}
	if len(bts) == 0 {
		return fmt.Errorf("empty slice")
	}
	m.incomingQueue.PushTail(bts)
	m.incomingNotifyChan <- 1
	return nil
}

func (m *BufferedFileWriter) Close(waitTime time.Duration) {
	m.stopped = true
	m.terminalSignal <- 1
	dur := time.Second * 10
	if waitTime > 0 {
		dur = waitTime
	}
	select {
	case <-m.terminalDone:
		return
	case <-time.After(dur):
		return
	}
}

func (m *BufferedFileWriter) Flush(waitTime time.Duration) {
	m.flushSignal <- 1
	dur := time.Second * 10
	if waitTime > 0 {
		dur = waitTime
	}
	select {
	case <-m.flushDone:
		return
	case <-time.After(dur):
		return
	}
}

func (m *BufferedFileWriter) flush() {
	buf := <-m.flushChan
	defer func() {
		m.freeChan <- 1
	}()
	defer buf.Reset()

	if buf.Len() == 0 {
		return
	}
	fn := filepath.Join(m.filePath, time.Now().Format(m.fileTimeMask))
	fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error-open file [%s] fail, %s", fn, err.Error())
	}
	defer fd.Close()
	writeBytes, err := fd.Write(buf.Bytes())
	if writeBytes != buf.Len() {
		fmt.Printf("error-write to file %s fail, %d bytes writen,expect %d.", fn, writeBytes, buf.Len())
	}
	fmt.Printf("%d bytes writed\n", writeBytes)
}

func (m *BufferedFileWriter) writeBuffer(bts []byte) {
	if bts == nil {
		return
	}
	if len(bts) == 0 {
		return
	}
	freeLen := m.bufferCap - m.bufferCurrent.Len()
	if freeLen < len(bts) {
		m.incomingQueue.PushHead(bts[freeLen:])
		m.incomingNotifyChan <- 1
		bts = bts[:freeLen]
	}
	m.bufferCurrent.Write(bts)
	if m.bufferCurrent.Len() >= m.bufferCap {
		m.notifyFlush()
		if m.bufferCurrent == &m.buffer1 {
			m.bufferCurrent = &m.buffer2
		} else {
			m.bufferCurrent = &m.buffer1
		}
	}
}

func (m *BufferedFileWriter) notifyFlush() {
	//waiting for flushing done
	<-m.freeChan
	//notify
	m.flushChan <- m.bufferCurrent
}

func (m *BufferedFileWriter) run() {
	for {
		select {
		case <-m.flushSignal:
			//process queued data
			for {
				data := m.incomingQueue.PopHead()
				if data != nil {
					m.writeBuffer(data.([]byte))
				} else {
					break
				}
			}

			m.notifyFlush()
			//waiting for flush done
			<- m.freeChan
			m.bufferCurrent.Reset()
			m.freeChan <- 1
		case <-m.terminalSignal:
			fmt.Println("terminated,shutdown")
			//process queued data
			for {
				data := m.incomingQueue.PopHead()
				if data != nil {
					m.writeBuffer(data.([]byte))
				} else {
					break
				}
			}

			//flush all data to file
			m.notifyFlush()
			//waiting for flushing done
			<-m.freeChan
			m.terminalDone <- 1
			return
		case <-m.incomingNotifyChan:
			data := m.incomingQueue.PopHead()
			if data != nil {
				m.writeBuffer(data.([]byte))
			}
		}
	}
}
