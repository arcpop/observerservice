
// +build windows

package main

import (
    "golang.org/x/sys/windows/svc"
	"fmt"
    "net"
)

//ObserverService is an empty struct to implement svc.Handler
type ObserverService struct {}

//Execute implements svc.Handler
func (s *ObserverService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
    const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
    changes <- svc.Status{State: svc.StartPending}
    if len(args) <= 1 {
        changes <- svc.Status{State: svc.StopPending}
        return true, 10
    }
    
    serverAddr, err := net.ResolveTCPAddr("tcp", args[1])
    if err != nil {
        changes <- svc.Status{State: svc.StopPending}
        return true, 20
    }
    
    incomingNotifications := make(chan Notification, NotificationQueueSize)
    outgoingNotifications := make(chan string, NotificationQueueSize)
    defer close(incomingNotifications)
    defer close(outgoingNotifications)
    
    go consolePrintNotifications(outgoingNotifications)
    //go sendNotifications(serverAddr, outgoingNotifications)
    
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
                go func (notification Notification) {
                    notification.Handle()
                    outgoingNotifications <- notification.Encode()
                } (nft)
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

