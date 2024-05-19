package wal

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	CommandBsize   = 3
	TimestampBsize = 8
	MaxWlogSize    = 4096
)
const (
	TypeJSON byte = iota + 1
	TypeString
	TypeInt
)

type WALManager struct {
	Fd                *os.File
	Log               []WALLogEntry
	LogStream         chan []WALLogEntry
	LatestCommitIndex uint64
	mu                sync.Mutex
}

func NewWALManager(filename string) *WALManager {
	fd, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	walManager := &WALManager{
		Fd:                fd,
		Log:               []WALLogEntry{},
		LogStream:         make(chan []WALLogEntry),
		mu:                sync.Mutex{},
		LatestCommitIndex: 0,
	}
	go walManager.LogListener()
	return walManager

}

func (wal *WALManager) LogListener() {
	for {
		select {
		case lo :=
			<-wal.LogStream:
			for _, entry := range lo {
				wal.mu.Lock()
				wal.AppendLog(entry)
				wal.mu.Unlock()
			}
		}
	}
}

func (walManager *WALManager) IncrementLatestCommitedIndex() uint64 {
	return atomic.AddUint64(&walManager.LatestCommitIndex, 1)
}

func (walManager *WALManager) AppendLog(entry WALLogEntry) {

	fb, err := encodeWALEntry(walManager, entry)
	if err != nil {
		return
	}
	_, err = decodeWALEntry(walManager, fb)
	if err != nil {
	}
	return
}

func encodeWALEntry(walManager *WALManager, entry WALLogEntry) ([]byte, error) {
	// Initialize buffer with the maximum size
	msg := bytes.NewBuffer(nil) // Encode timestamp
	// Create a buffered writer
	entry.Timestamp = uint64(time.Now().UnixNano())
	// Convert and write timestamp field to bytes
	err := binary.Write(msg, binary.BigEndian, entry.Timestamp)
	err = binary.Write(msg, binary.BigEndian, uint64(entry.LogIndex))
	binary.Write(msg, binary.BigEndian, uint8(len(entry.Key)))
	msg.WriteString(entry.Key) // Convert and write log index field to bytes
	// Write key field as bytes

	// Convert and write value field to bytes based on data type
	switch entry.Value.(type) {
	case string:
		binary.Write(msg, binary.BigEndian, uint8(0))
		binary.Write(msg, binary.BigEndian, uint8(len(entry.Value.(string))))
		msg.WriteString(entry.Value.(string))
		if err != nil {
			return nil, err
		}
	default:
		binary.Write(msg, binary.BigEndian, 1)
		encodedJson, _ := json.Marshal(entry)
		encodedJson = append(encodedJson, '\n')
		msg.Write(encodedJson)
	}
	// Write a newline to separate entries if needed
	// Flush the buffer to ensure all data is written to the file
	walManager.Fd.Write(msg.Bytes())
	return msg.Bytes(), nil
}

func decodeWALEntry(walManager *WALManager, data []byte) (*WALLogEntry, error) {
	//buf := make([]byte, 1)
	//bytesRead := []byte{}
	//for {
	//	_, err := walManager.Fd.Read(buf)
	//	if err != nil {
	//		log.Fatal("recieved error reading from file: ", err)
	//	}
	//	bytesRead = append(bytesRead, buf...)
	//	if buf[0] == '\n' {
	decodeBytesToWal(walManager, data)
	//}
	//}
	return nil, nil
}

func decodeBytesToWal(manager *WALManager, read []byte) *WALLogEntry {
	var timeStamp, Logindex uint64
	var encodingType uint8
	var klen, vlen uint8

	binary.Read(bytes.NewReader(read[:8]), binary.BigEndian, &timeStamp)
	binary.Read(bytes.NewReader(read[8:16]), binary.BigEndian, &Logindex)
	binary.Read(bytes.NewReader(read[16:17]), binary.BigEndian, &klen)
	keyBytes := make([]byte, klen)
	binary.Read(bytes.NewReader(read[17:17+klen]), binary.BigEndian, &keyBytes)
	fmt.Println(string(keyBytes))
	binary.Read(bytes.NewReader(read[17+klen:17+klen+1]), binary.BigEndian, &encodingType)
	binary.Read(bytes.NewReader(read[17+klen+1:17+klen+2]), binary.BigEndian, &vlen)
	vbytes := make([]byte, vlen)
	binary.Read(bytes.NewReader(read[17+klen+2:17+klen+2+vlen]), binary.BigEndian, &vbytes)
	fmt.Println(string(vbytes))

	log := &WALLogEntry{
		Key:       string(keyBytes),
		Value:     string(vbytes),
		Comm:      Command(encodingType),
		Timestamp: timeStamp,
		LogIndex:  Logindex,
	}
	fmt.Println(log)
	return log
}
func (walManager *WALManager) encodeTimeStamp(entry WALLogEntry) ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := binary.Write(buffer, binary.BigEndian, entry.Timestamp); err != nil {
		return nil, err
	}
	fullBytes := buffer.Bytes()
	if len(fullBytes) > 3 {
		fullBytes = fullBytes[len(fullBytes)-3:]
	}
	return fullBytes, nil
}
