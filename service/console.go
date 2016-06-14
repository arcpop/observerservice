package service

import (
	"fmt"
)


func consolePrintNotifications(s chan string) {
    fmt.Println(<-s)
}
