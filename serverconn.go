package main

import (
	"net"
)


func sendNotifications(serverAddr *net.TCPAddr, outgoingNotifications chan []byte)  {
    serverConn, err := net.DialTCP("tcp", nil, serverAddr)
    if err != nil {
        panic(err)
    }
    defer serverConn.Close()
    for nft := range outgoingNotifications {
        _, err = serverConn.Write(nft)
        if err != nil {
            panic(err)
        }
    }
}