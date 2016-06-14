package service

import (
	"net"
)


func sendNotifications(serverAddr *net.TCPAddr, outgoingNotifications chan string)  {
    serverConn, err := net.DialTCP("tcp", nil, serverAddr)
    if err != nil {
        panic(err)
    }
    defer serverConn.Close()
    for nft := range outgoingNotifications {
        _, err = serverConn.Write([]byte(nft))
        if err != nil {
            panic(err)
        }
    }
}