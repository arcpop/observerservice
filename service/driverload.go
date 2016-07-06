package service


import (
    "encoding/binary"
	"fmt"
	"encoding/json"
	"strconv"
)

//DriverLoadNotification represents a driver load notification
type DriverLoadNotification struct {
    NotificationHeader
    ImageBase uint64
    ImageSize uint64
    Truncated uint16
    ImageSigned bool
    ImagePath string
}

const (
    MinDriverLoadNotificationSize = NotificationHeaderSize + 24
)

//ParseFrom parses the driver load notification from a byte buffer
func (n *DriverLoadNotification) ParseFrom(b []byte) error {
    if len(b) < MinRegistryNotificationSize {
        return ErrParsingFailed
    }
    n.NotificationHeader.ParseFrom(b)
    n.ImageBase = binary.LittleEndian.Uint64(b[24:])
    n.ImageSize = binary.LittleEndian.Uint64(b[32:])
    flags := binary.LittleEndian.Uint32(b[40:])
    if (flags & 1) != 0 {
        n.ImageSigned = true
    }
    n.Truncated = binary.LittleEndian.Uint16(b[44:])
    n.ImagePath = decodeUnicodeByteBuffer(b[46:])
    return nil
}

//Encode encodes the notification to send it to the server
func (n *DriverLoadNotification) Encode() ([]byte, error) {
    return json.Marshal(*n)
}

func (n *DriverLoadNotification) Handle() {
    fmt.Println("Driver: " + n.ImagePath + " at 0x" + 
        strconv.FormatUint(n.ImageBase, 16))
}
