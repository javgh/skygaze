package skygazer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	interruptInterval      = 2 * time.Second
	apiEndpointTemplate    = "http://localhost:9980/skynet/metadata/%s"
	siaAgent               = "Sia-Agent"
	userAgentHeader        = "User-Agent"
	canonicalSkylinkHeader = "Skynet-Canonical-Skylink"
)

type (
	SkyGazer struct {
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
)

func New() *SkyGazer {
	return &SkyGazer{}
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
			verifiedSkylink, err := probe(line)
			if err != nil {
				log.Println(err)
			}
			fmt.Println(verifiedSkylink)
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

func probe(skylink string) (*VerifiedSkylink, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf(apiEndpointTemplate, skylink), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(userAgentHeader, siaAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	canonicalSkylink := resp.Header.Get(canonicalSkylinkHeader)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var metadata SkyfileMetadata
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return nil, err
	}

	return &VerifiedSkylink{
		CanonicalSkylink: canonicalSkylink,
		Metadata:         metadata,
	}, nil
}
