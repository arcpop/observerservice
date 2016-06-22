package service

import (
    "unicode/utf16"
    "encoding/binary"
	"fmt"
	"encoding/json"
)

//RegistryNotification represents a registry notification
type RegistryNotification struct {
    NotificationHeader
    RegistryAction uint16
    Truncated uint16
    RegistryPath string
}

const (
    MinRegistryNotificationSize = NotificationHeaderSize + 6
)

//ParseFrom parses the registry notification from a byte buffer
func (n *RegistryNotification) ParseFrom(b []byte) error {
    if len(b) < MinRegistryNotificationSize {
        return ErrParsingFailed
    }
    n.NotificationHeader.ParseFrom(b)
    n.RegistryAction = binary.LittleEndian.Uint16(b[24:26])
    n.Truncated = binary.LittleEndian.Uint16(b[26:28])
    n.RegistryPath = decodeUnicodeByteBuffer(b[28:])
    return nil
}

//Encode encodes the notification to send it to the server
func (n *RegistryNotification) Encode() ([]byte, error) {
    return json.Marshal(*n)
}

//Handle should perform actions upon receiving this type of notification
func (n *RegistryNotification) Handle() {
    fmt.Println("Registry: (" + n.NotificationHeader.CurrentProcess + ")" + n.RegistryPath)
}

func decodeUnicodeByteBuffer(b []byte) string {
    buf := make([]uint16, len(b) / 2)
    count := (len(b) - 1) / 2
    for i := 0; i < count; i++ {
        c := binary.LittleEndian.Uint16(b[2 * i:])
        if c == 0 {
            s := string(utf16.Decode(buf[:i]))
            return s
        }
        buf[i] = c
    }
    s := string(utf16.Decode(buf))
    return s
}