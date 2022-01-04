package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/micro/cli/v2"
	"golang.org/x/net/proxy"
)

func main() {
	app := cli.NewApp()

	app.Usage = "Connect tcp over socks5"

	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "local listen port",
			Value:   12345,
		},
		&cli.StringFlag{
			Name:    "socks5",
			Aliases: []string{"x"},
			Usage:   "socks5 server",
			Value:   "127.0.0.1:1080",
		},
		&cli.StringFlag{
			Name:     "target",
			Aliases:  []string{"t"},
			Usage:    "target address",
			Required: true,
		},
	}

	app.Action = func(c *cli.Context) error {
		port := c.Int("port")
		socks5 := c.String("socks5")
		target := c.String("target")

		addr := fmt.Sprintf(":%d", port)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Printf("listen fail: %v\n", err)
			return err
		}

		socks5dialer, err := proxy.SOCKS5("tcp", socks5, nil, nil)
		if err != nil {
			fmt.Printf("new socks5 dialer fail: %v\n", err)
			return err
		}

		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Printf("accept fail: %v\n", err)
				continue
			}

			go handleConnection(socks5dialer.Dial, conn, target)
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("run fail: %v", err)
	}
}

func handleConnection(dialer func(network, address string) (net.Conn, error), conn net.Conn, addr string) {
	defer conn.Close()

	remote, err := dialer("tcp", addr)
	if err != nil {
		fmt.Printf("dial %s fail: %v\n", addr, err)
	}
	defer remote.Close()

	go io.Copy(conn, remote)
	io.Copy(remote, conn)
}
