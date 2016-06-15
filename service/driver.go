
// +build windows

package service

import (
    "encoding/binary"
    "strconv"
    "errors"
	"fmt"
	"os"
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
    Encode() ([]byte, error)
}

type DriverListener struct {
    handle *os.File
    doClose chan bool
    notificationsChan chan Notification
}

func createDriverListener(driverName string, notifications chan Notification) (*DriverListener, error) {
    fd, err := os.OpenFile("\\\\.\\" + driverName, os.O_RDONLY, 0)
    
    if err != nil {
        println(err.Error())
        return nil, err
    }
    return &DriverListener{handle: fd, doClose: make(chan bool), notificationsChan: notifications}, nil
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
                dl.handle.Close()
                return
            default:
                n, err := dl.ReadMessage()
                if err != nil {
                    if err != ErrInvalidNotificationType {
                        fmt.Println(err)
                    }
                    continue
                }
                dl.notificationsChan <- n
        }
    }
}

//NotificationHeader is the header which is sent for all notifications
type NotificationHeader struct {
    NotificationType uint32
    Reaction uint32
    CurrentProcessID uint64
    CurrentThreadID uint64
}

//ParseFrom parses the notification header from a byte buffer
func (n NotificationHeader) ParseFrom(b []byte) error {
    if len(b) < 24 {
        return ErrParsingFailed
    }
    n.NotificationType = binary.BigEndian.Uint32(b[0:4])
    n.Reaction = binary.BigEndian.Uint32(b[4:8])
    n.CurrentProcessID = binary.BigEndian.Uint64(b[8:16])
    n.CurrentThreadID = binary.BigEndian.Uint64(b[16:24])
    return nil
}

//EncodeHeader is used so that NotificationHeader does not implement the Notification interface
func (n NotificationHeader) EncodeHeader() []byte {
    b := make([]byte, 24, DefaultEncodingSize)
    binary.BigEndian.PutUint32(b[0:4], n.NotificationType)
    binary.BigEndian.PutUint32(b[4:8], n.Reaction)
    binary.BigEndian.PutUint64(b[8:16], n.CurrentProcessID)
    binary.BigEndian.PutUint64(b[16:24], n.CurrentThreadID)
    return b
}

//ReadMessage reads a single notification from the driver
func (dl *DriverListener) ReadMessage() (Notification, error) {
    buffer := make([]byte, 20000)
    n, err := dl.handle.Read(buffer[:])
    if err != nil {
        fmt.Printf("ReadMessage(): Read error: %+v\n", err)
        return nil, err
    }
    if n < 4 {
        fmt.Println("ReadMessage(): Not enough bytes returned: ", strconv.FormatInt(int64(n), 10))
        return nil, ErrParsingFailed
    }

    notificationType := binary.LittleEndian.Uint32(buffer[0:4])
    switch notificationType {
    case NotificationRegistry:
        rn := &RegistryNotification{}
        err = rn.ParseFrom(buffer[0:n])
        if err != nil {
            return nil, err
        }
        return rn, nil
    }
    return nil, ErrInvalidNotificationType
}