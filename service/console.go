package service

import (
	"fmt"
)


func consolePrintNotifications(s chan []byte) {
    fmt.Println("[Emulating server on console]")
    for {
        _ = string(<-s)
        //fmt.Println("Server receives: ", str)
    }
}
