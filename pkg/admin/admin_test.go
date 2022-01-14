package admin_test

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/admin/test"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"testing"
)

func TestAdminHandlers(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	handlers := admin.LoadHandlers("localhost:8885", new(admin_test.MockClient))
	server := web_support.Create(listener.Addr().String(), handlers, web_support.Options{})
	go web_support.Start(server, listener)
	web_support.WaitForHealthy(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	noFollowClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	redirect, _ := noFollowClient.Get(fmt.Sprintf("http://%s", server.Addr))
	assert.Equal(t, http.StatusPermanentRedirect, redirect.StatusCode)
	assert.Equal(t, string(redirect.Header["Location"][0]), "/integrations")

	web_support.Stop(server)
}
