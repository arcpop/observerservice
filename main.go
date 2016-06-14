
// +build windows

package main



import (
    "github.com/arcpop/observerservice/service"
    "golang.org/x/sys/windows/svc"
	"os/signal"
	"os"
	"fmt"
)

const (
    //ServiceName is the name the service will use
    ServiceName = "observerservice"

)

func main()  {
    /*isInteractiveSession, err := svc.IsAnInteractiveSession()
    if err != nil {
        panic(err)
    }
    if isInteractiveSession {
        panic("Service can't run in an interactive session")
    }
    
    service := &service.ObserverService{}
    
    err = svc.Run(ServiceName, service)
    if err != nil {
        return
    }*/
    requests := make(chan svc.ChangeRequest)
    changes  := make(chan svc.Status)
    service := &service.ObserverService{}
    go func () {
        _, err := service.Execute([]string{"observerservice.exe", "console",}, requests, changes)
        panic(err)
    } ()

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    for {
        select {
            case <- c:
                fmt.Println("Ctrl+C")
                requests <- svc.ChangeRequest{
                    Cmd: svc.Stop,
                }
                return
            case chng := <- changes:
                fmt.Println("Status: ", chng)
        }
    }
}