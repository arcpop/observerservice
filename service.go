
// +build windows

package main

import (
    "golang.org/x/sys/windows/svc"
	"fmt"
)

//ObserverService is an empty struct to implement svc.Handler
type ObserverService struct {}

//Execute implements svc.Handler
func (s *ObserverService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
    const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
    changes <- svc.Status{State: svc.StartPending}
    
    incomingNotifications := make(chan Notification, NotificationQueueSize)
    outgoingNotifications := make(chan []byte, NotificationQueueSize)
    
    driverListener, err := createDriverListener(DriverName, incomingNotifications)
    if err != nil {
        changes <- svc.Status{State: svc.Stopped}
        return false, 0
    }
    defer driverListener.Close()
    
    changes <- svc.Status{State: svc.Running}
    for {
        select {
            case nft := <- incomingNotifications:
                nft.Handle()
                outgoingNotifications <- nft.Encode()
            case req := <- r:
                switch req.Cmd {
                case svc.Stop, svc.Shutdown:
                    changes <- svc.Status{State: svc.StopPending}
                    return
                default:
                    fmt.Printf("Failed command: %v", req.Cmd)
                }
        }
    }
}

