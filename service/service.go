
// +build windows

package service

import (
    "golang.org/x/sys/windows/svc"
	"fmt"
    "net"
)

const (
    //NotificationQueueSize is the number of non blocking entries 
    //in the notification queue (implemented over a channel)
    NotificationQueueSize = 500

    //DriverName is the dos name of the driver
    DriverName = "Observer"
)

//ObserverService is an empty struct to implement svc.Handler
type ObserverService struct {}

//Execute implements svc.Handler
func (s *ObserverService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
    const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
    changes <- svc.Status{State: svc.StartPending}

    if len(args) <= 1 {
        println("No args")
        changes <- svc.Status{State: svc.StopPending}
        return true, 10
    }
    
    
    incomingNotifications := make(chan Notification, NotificationQueueSize)
    outgoingNotifications := make(chan string, NotificationQueueSize)
    defer close(incomingNotifications)
    defer close(outgoingNotifications)
    
    if (args[1] != "console") {
        serverAddr, err := net.ResolveTCPAddr("tcp", args[1])
        if err != nil {
            println("Failed to resolve server addr")
            changes <- svc.Status{State: svc.StopPending}
            return true, 20
        }
        go sendNotifications(serverAddr, outgoingNotifications)
    } else {
        go consolePrintNotifications(outgoingNotifications)
    }

    driverListener, err := createDriverListener(DriverName, incomingNotifications)
    if err != nil {
        println("Failed to create DriverListener")
        changes <- svc.Status{State: svc.Stopped}
        return false, 0
    }
    go driverListener.ListenForNotifications()

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

