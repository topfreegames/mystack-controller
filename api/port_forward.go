// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>
package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/topfreegames/mystack-controller/extensions"
	"github.com/topfreegames/mystack-controller/models"
)

func CopyConn(dst, src net.Conn) {
	defer dst.Close()
	defer src.Close()

	io.Copy(dst, src)
}

func Proxy(remoteConn, clientConn net.Conn, logger logrus.FieldLogger) {
	logger.Printf("Proxied to '%s'", remoteConn.RemoteAddr())
	go CopyConn(clientConn, remoteConn)
	go CopyConn(remoteConn, clientConn)
}

func Read(conn net.Conn, logger logrus.FieldLogger) (string, string, error) {
	buf, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return "", "", fmt.Errorf("error reading handshake: %s", err)
	}

	obj := make(map[string]interface{})
	err = json.Unmarshal(buf, &obj)
	if err != nil {
		return "", "", err
	}

	logger.Debug("received message from tcp socket")

	return obj["token"].(string), obj["service"].(string), nil
}

func (a *App) listenTCP(url string) {
	listener, err := net.Listen("tcp", url)
	if err != nil {
		a.Logger.Fatalf("Failed to setup listener: %v", err)
	}

	for {
		conn, err := listener.Accept()
		a.Logger.Infof("accepted new connection")
		if err != nil {
			a.Logger.Fatalf("error: failed to accept listener: %v", err)
		}

		accessToken, service, err := Read(conn, a.Logger)
		if err != nil {
			fmt.Fprintf(conn, "error reading handshake message: %s", err)
			continue
		}

		token, err := extensions.Token(accessToken, a.DB)
		if err != nil {
			fmt.Fprintf(conn, "error: connection was not authenticated")
			continue
		}
		a.Logger.Infof("validated token")

		email, _, err := extensions.Authenticate(token, &models.OSCredentials{}, a.DB)
		if err != nil {
			fmt.Fprintf(conn, "connection was not authenticated: %s", err)
			continue
		}
		username := usernameFromEmail(email)
		if !a.verifyEmailDomain(email) {
			conn.Write([]byte("unauthorized email"))
			continue
		}

		a.Logger.Infof("proxying application for %s", email)

		conn.Write([]byte("successfull authentication"))

		port, err := models.ServicePort(a.Clientset, service, username)
		if err != nil {
			fmt.Fprintf(conn, "service doesn't exist: %s", err)
			continue
		}

		remoteAddr := fmt.Sprintf("%s.mystack-%s:%d", service, username, port)
		remoteConn, err := net.Dial("tcp", remoteAddr)
		if err != nil {
			a.Logger.Fatalf("Dial failed: %v", err)
		}

		a.Logger.Infof("accepted connection, proxying to %s", remoteAddr)
		Proxy(remoteConn, conn, a.Logger)
	}
}
