
// +build windows

package main



import (
    "golang.org/x/sys/windows/svc"
)

const (
    //ServiceName is the name the service will use
    ServiceName = "observerservice"

    //DriverName is the dos name of the driver
    DriverName = "Observer"
    
    //NotificationQueueSize is the number of non blocking entries 
    //in the notification queue (implemented over a channel)
    NotificationQueueSize = 500
)

func main()  {
    isInteractiveSession, err := svc.IsAnInteractiveSession()
    if err != nil {
        panic(err)
    }
    if isInteractiveSession {
        panic("Service can't run in an interactive session")
    }
    
    service := &ObserverService{}
    
    err = svc.Run(ServiceName, service)
    if err != nil {
        return
    }
}