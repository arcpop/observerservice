package service

import (
	"encoding/json"
	"fmt"
	"encoding/binary"
	"golang.org/x/sys/windows"
	"unsafe"
	"errors"
)

//ProcessCreationNotification is the go type of ProcessCreated notification
type ProcessCreationNotification struct {
    NotificationHeader
    NewProcessID uint64
    ParentProcessID uint64
    CreatingThreadID uint64
    CreatingProcessID uint64
    ProcessName string
}

const (
    MinProcessCreationNotificationSize = NotificationHeaderSize + 36
)

var (
    ErrProcessNotFound = errors.New("Process entry not found")
)

func (n *ProcessCreationNotification) ParseFrom(b []byte) error {
    if len(b) < MinRegistryNotificationSize {
        return ErrParsingFailed
    }
    n.NotificationHeader.ParseFrom(b)
    n.NewProcessID = binary.LittleEndian.Uint64(b[24:])
    n.ParentProcessID = binary.LittleEndian.Uint64(b[32:])
    n.CreatingThreadID = binary.LittleEndian.Uint64(b[40:])
    n.CreatingProcessID = binary.LittleEndian.Uint64(b[48:])
    truncated := binary.LittleEndian.Uint16(b[56:])
    if truncated == 0 {
        n.ProcessName = decodeUnicodeByteBuffer(b[58:])
    } else {
        name, err := findProcessNameByID(n.NewProcessID)
        if err != nil {
            //Use truncated name anyway
            n.ProcessName = decodeUnicodeByteBuffer(b[58:])
        } else {
            n.ProcessName = name
        }
    }
    return nil
}

func (n *ProcessCreationNotification) Encode() ([]byte, error) {
    return json.Marshal(*n)
}

func (n *ProcessCreationNotification) Handle() {
    fmt.Println("Handle process created: " + n.ProcessName)
}


func findProcessNameByID(pid uint64) (string, error) {
    var processEntry windows.ProcessEntry32
    handle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
    if err != nil {
        return "", err
    }
    defer windows.CloseHandle(handle)
    processEntry.Size = uint32(unsafe.Sizeof(processEntry))
    for err = windows.Process32First(handle, &processEntry); 
        err != nil;
        err = windows.Process32Next(handle, &processEntry) {
        if processEntry.ProcessID == uint32(pid) {
            return windows.UTF16ToString(processEntry.ExeFile[:]), nil
        }
    }
    if err != nil {
        return "", err
    }
    return "", ErrProcessNotFound
}