// client.go
package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "os"
)

func main() {
    if len(os.Args) < 5 {
        log.Fatalf("Usage: client <server-addr> <subdomain> <local-port> <token>")
    }

    serverAddr := os.Args[1]
    subdomain := os.Args[2]
    localPort := os.Args[3]
    token := os.Args[4]

    for {
        conn, err := net.Dial("tcp", serverAddr)
        if err != nil {
            log.Println("Failed to connect to server:", err)
            continue
        }

        fmt.Fprintf(conn, "%s:%s\n", subdomain, token)

        log.Printf("Connected to server as %s", subdomain)

        go func() {
            for {
                req, err := http.ReadRequest(bufio.NewReader(conn))
                if err != nil {
                    log.Println("Connection closed")
                    conn.Close()
                    return
                }

                url := fmt.Sprintf("http://localhost:%s%s", localPort, req.URL.String())
                newReq, _ := http.NewRequest(req.Method, url, req.Body)
                newReq.Header = req.Header

                resp, err := http.DefaultClient.Do(newReq)
                if err != nil {
                    log.Println("Local request failed:", err)
                    continue
                }

                resp.Write(conn)
            }
        }()

        io.Copy(io.Discard, conn)
    }
}