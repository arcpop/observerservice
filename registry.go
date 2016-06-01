package main

import (
    "unicode/utf16"
    "encoding/binary"
	"fmt"
)

//RegistryNotification represents a registry notification
type RegistryNotification struct {
    NotificationHeader
    registryAction uint16
    truncated uint16
    registryPath string
}

//ParseFrom parses the registry notification from a byte buffer
func (n RegistryNotification) ParseFrom(b []byte) error {
    if len(b) < 29 {
        return ErrParsingFailed
    }
    n.NotificationHeader.ParseFrom(b)
    n.registryAction = binary.BigEndian.Uint16(b[24:26])
    n.truncated = binary.BigEndian.Uint16(b[26:28])
    n.registryPath = decodeUnicodeByteBuffer(b[28:])
    return nil
}

//Encode encodes the notification to send it to the server
func (n RegistryNotification) Encode() string {
    return fmt.Sprintf("[Registry %d:%d] -> %s (%d)", 
    n.currentProcessID, 
    n.currentThreadID, 
    n.registryPath, 
    n.registryAction)
}

//Handle should perform actions upon receiving this type of notification
func (n RegistryNotification) Handle() {
    println("Registry: " + n.registryPath)
}

func decodeUnicodeByteBuffer(b []byte) string {
    buf := make([]uint16, len(b) / 2)
    count := (len(b) - 1) / 2
    for i := 0; i < count; i++ {
        if b[2 * i] == 0 && b[2 * i + 1] == 0 {
            return string(utf16.Decode(buf[:i - 1]))
        }
        buf[i] = binary.BigEndian.Uint16(b[2 * i:])
    }
    return string(utf16.Decode(buf))
}