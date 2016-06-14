
// +build windows

package service

import (
    "encoding/binary"
    "strconv"
    "errors"
    "golang.org/x/sys/windows"
	"fmt"
	"syscall"
)

const (
    DefaultEncodingSize = 1024
)
const (
    NotificationRegistry = 1
    NotificationThreadCreated = 2
    NotificationProcessCreated = 3
    NotificationModuleLoad = 4
    NotificationDriverLoad = 5
)

var (
    ErrParsingFailed = errors.New("Failed to parse")
    ErrInvalidNotificationType = errors.New("Received an invalid notification type")
)

type Notification interface {
    Handle()
    Encode() string
}

type DriverListener struct {
    handle windows.Handle
    doClose chan bool
}

func createDriverListener(driverName string, notifications chan Notification) (*DriverListener, error) {
    fd, err := windows.CreateFile(
        windows.StringToUTF16Ptr("\\\\.\\" + driverName), 
        windows.GENERIC_READ,
        windows.FILE_SHARE_READ,
        nil,
        windows.OPEN_EXISTING,
        windows.FILE_FLAG_OVERLAPPED,
        0)

    if err != nil {
        println(err.Error())
        return nil, err
    }
    return &DriverListener{handle: fd, doClose: make(chan bool)}, nil
}

//Close closes the associated driver handle
func (dl *DriverListener) Close() error {
    dl.doClose <- true
    return nil
}

//ListenForNotifications listens for notifications from the driver
func (dl *DriverListener) ListenForNotifications() {
    for {
        select {
            case <- dl.doClose:
                windows.Close(dl.handle)
                return
            default:
                dl.ReadMessage()
        }
    }
}

//NotificationHeader is the header which is sent for all notifications
type NotificationHeader struct {
    notificationType uint32
    reaction uint32
    currentProcessID uint64
    currentThreadID uint64
}

//ParseFrom parses the notification header from a byte buffer
func (n NotificationHeader) ParseFrom(b []byte) error {
    if len(b) < 24 {
        return ErrParsingFailed
    }
    n.notificationType = binary.BigEndian.Uint32(b[0:4])
    n.reaction = binary.BigEndian.Uint32(b[4:8])
    n.currentProcessID = binary.BigEndian.Uint64(b[8:16])
    n.currentThreadID = binary.BigEndian.Uint64(b[16:24])
    return nil
}

//EncodeHeader is used so that NotificationHeader does not implement the Notification interface
func (n NotificationHeader) EncodeHeader() []byte {
    b := make([]byte, 24, DefaultEncodingSize)
    binary.BigEndian.PutUint32(b[0:4], n.notificationType)
    binary.BigEndian.PutUint32(b[4:8], n.reaction)
    binary.BigEndian.PutUint64(b[8:16], n.currentProcessID)
    binary.BigEndian.PutUint64(b[16:24], n.currentThreadID)
    return b
}

//ReadMessage reads a single notification from the driver
func (dl *DriverListener) ReadMessage() (Notification, error) {
    buffer := make([]byte, 20000)
    var done uint32
    event, err := windows.CreateEvent(nil, 1, 0, nil)
    if err != nil {
        fmt.Printf("ReadMessage(): windows.CreateEvent error: %+v\n", err)
        return nil, err
    }
    overlapped := windows.Overlapped{
        HEvent: event,
    }
    err = windows.ReadFile(dl.handle, buffer[:], &done, &overlapped)
    if err != nil {
        errn, ok := err.(syscall.Errno)
        if !ok || errn != windows.ERROR_IO_PENDING {
            fmt.Printf("ReadMessage(): syscall.Read error: %+v\n", err)
            return nil, err
        }
    }
    ev, err := windows.WaitForSingleObject(event, windows.INFINITE)
    fmt.Printf("ReadMessage(): windows.WaitForSingleObject error: %+v %T\n", err, err)
    fmt.Printf("ReadMessage(): windows.WaitForSingleObject event: %+v %T\n", ev, ev)
    
    if done < 4 {
        fmt.Println("ReadMessage(): Not enough bytes returned: ", strconv.FormatInt(int64(done), 10))
        return nil, ErrParsingFailed
    }
    notificationType := binary.BigEndian.Uint32(buffer[0:4])
    switch notificationType {
    case NotificationRegistry:
        rn := RegistryNotification{}
        rn.ParseFrom(buffer[0:])
        return rn, nil
        
    }
    return nil, ErrInvalidNotificationType
}