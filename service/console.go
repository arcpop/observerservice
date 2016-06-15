package service

import (
	"fmt"
)


func consolePrintNotifications(s chan []byte) {
    fmt.Println("[Emulating server on console]")
    for {
        fmt.Println("Server receives: ", string(<-s))
    }
}
