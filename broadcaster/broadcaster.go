package broadcaster

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/javgh/skygaze/skygazer"
)

const (
	interruptInterval = 2 * time.Second
	template          = "https://siasky.net/%s | %s\n"
	cacheExpiration   = 30 * time.Second
	cacheInterval     = 5 * time.Second
	connectionTimeout = 30 * time.Second
)

type (
	Broadcaster struct {
		mutex       *sync.Mutex
		connections map[string]net.Conn
		cache       *cache.Cache
	}
)

func New() *Broadcaster {
	return &Broadcaster{
		mutex:       &sync.Mutex{},
		connections: make(map[string]net.Conn),
		cache:       cache.New(cacheExpiration, cacheInterval),
	}
}

func (b *Broadcaster) Serve(ctx context.Context, address string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}

	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	for ctx.Err() == nil {
		// Wake up from Accept() periodically to
		// check if we need to shutdown the server.
		err = ln.SetDeadline(time.Now().Add(interruptInterval))
		if err != nil {
			return err
		}

		conn, err := ln.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			return err
		}

		b.mutex.Lock()
		b.connections[conn.RemoteAddr().String()] = conn
		b.mutex.Unlock()
	}

	err = ln.Close()
	if err != nil {
		return err
	}

	return nil
}

func (b *Broadcaster) Broadcast(link skygazer.VerifiedSkylink) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	line := fmt.Sprintf(template, link.CanonicalSkylink, link.Metadata.Filename)
	for remoteAddr, conn := range b.connections {
		key := remoteAddr + "|" + link.CanonicalSkylink
		_, cached := b.cache.Get(key)
		if !cached {
			b.cache.Set(key, true, cache.DefaultExpiration)

			err := conn.SetDeadline(time.Now().Add(connectionTimeout))
			if err != nil {
				_ = conn.Close()
				delete(b.connections, remoteAddr)
			}

			_, err = conn.Write([]byte(line))
			if err != nil {
				_ = conn.Close()
				delete(b.connections, remoteAddr)
			}
		}
	}
}
