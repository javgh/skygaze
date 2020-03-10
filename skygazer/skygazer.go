package skygazer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	interruptInterval      = 2 * time.Second
	apiEndpointTemplate    = "http://localhost:9980/skynet/metadata/%s"
	siaAgent               = "Sia-Agent"
	userAgentHeader        = "User-Agent"
	canonicalSkylinkHeader = "Skynet-Canonical-Skylink"
	cacheExpiration        = 60 * time.Minute
	cacheInterval          = time.Minute
)

type (
	SkyGazer struct {
		cache *cache.Cache
	}

	SkyfileMetadata struct {
		Mode     os.FileMode     `json:"mode,omitempty"`
		Filename string          `json:"filename,omitempty"`
		Subfiles SkyfileSubfiles `json:"subfiles,omitempty"`
	}

	SkyfileSubfiles map[string]SkyfileSubfileMetadata

	SkyfileSubfileMetadata struct {
		Mode        os.FileMode `json:"mode,omitempty"`
		Filename    string      `json:"filename,omitempty"`
		ContentType string      `json:"contenttype,omitempty"`
		Offset      uint64      `json:"offset,omitempty"`
		Len         uint64      `json:"len,omitempty"`
	}

	VerifiedSkylink struct {
		CanonicalSkylink string
		Metadata         SkyfileMetadata
	}

	maybeVerifiedSkylink struct {
		successfullyVerified bool
		verifiedSkylink      *VerifiedSkylink
	}
)

func New() *SkyGazer {
	return &SkyGazer{
		cache: cache.New(cacheExpiration, cacheInterval),
	}
}

func (sg *SkyGazer) Listen(ctx context.Context, socketPath string) error {
	unixAddr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return err
	}

	ln, err := net.ListenUnix("unix", unixAddr)
	if err != nil {
		return err
	}

	for ctx.Err() == nil {
		// Wake up from Accept() periodically to
		// check if we need to shutdown the server.
		ln.SetDeadline(time.Now().Add(interruptInterval))
		conn, err := ln.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			return err
		}

		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			line := scanner.Text()

			cacheEntry, cached := sg.cache.Get(line)
			var mVerifiedSkylink maybeVerifiedSkylink
			if cached {
				mVerifiedSkylink = cacheEntry.(maybeVerifiedSkylink)
			} else {
				mVerifiedSkylink = probe(line)
				sg.cache.Set(line, mVerifiedSkylink, cache.DefaultExpiration)
			}

			fmt.Println("cached:", cached)
			fmt.Println(mVerifiedSkylink)
			fmt.Println(mVerifiedSkylink.verifiedSkylink)
		}

		err = conn.Close()
		if err != nil {
			return err
		}
	}

	err = ln.Close()
	if err != nil {
		return err
	}

	return nil
}

func probe(skylink string) maybeVerifiedSkylink {
	link := maybeVerifiedSkylink{}

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf(apiEndpointTemplate, skylink), nil)
	if err != nil {
		return link
	}

	req.Header.Add(userAgentHeader, siaAgent)
	resp, err := client.Do(req)
	if err != nil {
		return link
	}
	defer resp.Body.Close()

	canonicalSkylink := resp.Header.Get(canonicalSkylinkHeader)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return link
	}

	var metadata SkyfileMetadata
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return link
	}

	link.successfullyVerified = true
	link.verifiedSkylink = &VerifiedSkylink{
		CanonicalSkylink: canonicalSkylink,
		Metadata:         metadata,
	}
	return link
}
