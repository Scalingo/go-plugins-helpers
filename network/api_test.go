package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-plugins-helpers/sdk"
	"github.com/sirupsen/logrus"

	"github.com/Scalingo/go-utils/logger"
)

type TestDriver struct {
	Driver
}

func (t *TestDriver) GetCapabilities(_ context.Context) (*CapabilitiesResponse, error) {
	return &CapabilitiesResponse{Scope: LocalScope, ConnectivityScope: GlobalScope}, nil
}

func (t *TestDriver) CreateNetwork(_ context.Context, r *CreateNetworkRequest) error {
	return nil
}

func (t *TestDriver) DeleteNetwork(_ context.Context, r *DeleteNetworkRequest) error {
	return nil
}

func (t *TestDriver) CreateEndpoint(_ context.Context, r *CreateEndpointRequest) (*CreateEndpointResponse, error) {
	return &CreateEndpointResponse{}, nil
}

func (t *TestDriver) DeleteEndpoint(_ context.Context, r *DeleteEndpointRequest) error {
	return nil
}

func (t *TestDriver) Join(_ context.Context, r *JoinRequest) (*JoinResponse, error) {
	return &JoinResponse{}, nil
}

func (t *TestDriver) Leave(_ context.Context, r *LeaveRequest) error {
	return nil
}

func (t *TestDriver) ProgramExternalConnectivity(_ context.Context, r *ProgramExternalConnectivityRequest) error {
	i := r.Options["com.docker.network.endpoint.exposedports"]
	epl, ok := i.([]interface{})
	if !ok {
		return fmt.Errorf("invalid data in request: %v (%T)", i, i)
	}
	ep, ok := epl[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid data in request: %v (%T)", epl[0], epl[0])
	}
	if ep["Proto"].(float64) != 6 || ep["Port"].(float64) != 70 {
		return fmt.Errorf("Unexpected exposed ports in request: %v", ep)
	}
	return nil
}

func (t *TestDriver) RevokeExternalConnectivity(_ context.Context, r *RevokeExternalConnectivityRequest) error {
	return nil
}

type ErrDriver struct {
	Driver
}

func (e *ErrDriver) GetCapabilities(_ context.Context) (*CapabilitiesResponse, error) {
	return nil, fmt.Errorf("I CAN HAZ ERRORZ")
}

func (e *ErrDriver) CreateNetwork(_ context.Context, r *CreateNetworkRequest) error {
	return fmt.Errorf("I CAN HAZ ERRORZ")
}

func (e *ErrDriver) DeleteNetwork(_ context.Context, r *DeleteNetworkRequest) error {
	return errors.New("I CAN HAZ ERRORZ")
}

func (e *ErrDriver) CreateEndpoint(_ context.Context, r *CreateEndpointRequest) (*CreateEndpointResponse, error) {
	return nil, errors.New("I CAN HAZ ERRORZ")
}

func (e *ErrDriver) DeleteEndpoint(_ context.Context, r *DeleteEndpointRequest) error {
	return errors.New("I CAN HAZ ERRORZ")
}

func (e *ErrDriver) Join(_ context.Context, r *JoinRequest) (*JoinResponse, error) {
	return nil, errors.New("I CAN HAZ ERRORZ")
}

func (e *ErrDriver) Leave(_ context.Context, r *LeaveRequest) error {
	return errors.New("I CAN HAZ ERRORZ")
}

func callURL(url string) {
	c := http.Client{
		Timeout: 10 * time.Millisecond,
	}
	res := make(chan interface{}, 1)
	go func() {
		for {
			_, err := c.Get(url)
			if err == nil {
				res <- nil
			}
		}
	}()

	select {
	case <-res:
		return
	case <-time.After(5 * time.Second):
		fmt.Printf("Timeout connecting to %s\n", url)
		os.Exit(1)
	}
}

func TestMain(m *testing.M) {
	d := &TestDriver{}
	h1 := NewHandler(logrus.FieldLogger(logger.Default()), d)
	go func() {
		err := h1.ServeTCP("test", "localhost:32234", "", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up the TCP server: %v\n", err)
			os.Exit(-1)
		}
	}()

	e := &ErrDriver{}
	h2 := NewHandler(logrus.FieldLogger(logger.Default()), e)
	go func() {
		err := h2.ServeTCP("test", "localhost:32567", "", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up the TCP server: %v\n", err)
			os.Exit(-1)
		}
	}()

	// Test that the ServeTCP is ready for use.
	callURL("http://localhost:32234/Plugin.Activate")
	callURL("http://localhost:32567/Plugin.Activate")

	os.Exit(m.Run())
}

func TestActivate(t *testing.T) {
	response, err := http.Get("http://localhost:32234/Plugin.Activate")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	if string(body) != manifest+"\n" {
		t.Fatalf("Expected %s, got %s\n", manifest+"\n", string(body))
	}
}

func TestCapabilitiesExchange(t *testing.T) {
	response, err := http.Get("http://localhost:32234/NetworkDriver.GetCapabilities")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	expected := `{"Scope":"local","ConnectivityScope":"global"}`
	if string(body) != expected+"\n" {
		t.Fatalf("Expected %s, got %s\n", expected+"\n", string(body))
	}
}

func TestCreateNetworkSuccess(t *testing.T) {
	request := `{"NetworkID":"d76cfa51738e8a12c5eca71ee69e9d65010a4b48eaad74adab439be7e61b9aaf","Options":{"com.docker.network.generic":{}},"IPv4Data":[{"AddressSpace":"","Gateway":"172.18.0.1/16","Pool":"172.18.0.0/16"}],"IPv6Data":[]}`

	response, err := http.Post("http://localhost:32234/NetworkDriver.CreateNetwork",
		sdk.DefaultContentTypeV1_1,
		strings.NewReader(request),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	if response.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d\n", response.StatusCode)
	}
	if string(body) != "{}\n" {
		t.Fatalf("Expected %s, got %s\n", "{}\n", string(body))
	}
}

func TestCreateNetworkError(t *testing.T) {
	request := `{"NetworkID":"d76cfa51738e8a12c5eca71ee69e9d65010a4b48eaad74adab439be7e61b9aaf","Options":{"com.docker.network.generic":    {}},"IPv4Data":[{"AddressSpace":"","Gateway":"172.18.0.1/16","Pool":"172.18.0.0/16"}],"IPv6Data":[]}`
	response, err := http.Post("http://localhost:32567/NetworkDriver.CreateNetwork",
		sdk.DefaultContentTypeV1_1,
		strings.NewReader(request))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected 500, got %d\n", response.StatusCode)
	}

	expectedBody := "{\"Err\":\"I CAN HAZ ERRORZ\"}\n"
	if string(body) != expectedBody {
		t.Fatalf("Expected %s, got %s\n", expectedBody, string(body))
	}
}

func TestProgramExternalConnectivity(t *testing.T) {
	request := `{"NetworkID":"d76cfa51738e8a12c5eca71ee69e9d65010a4b48eaad74adab439be7e61b9aaf","EndpointID":"abccfa51738e8a12c5eca71ee69e9d65010a4b48eaad74adab439be7e61b9aaf","Options":{"com.docker.network.endpoint.exposedports":[{"Proto":6,"Port":70}],"com.docker.network.portmap":[{"Proto":6,"IP":"","Port":70,"HostIP":"","HostPort":7000,"HostPortEnd":7000}]}}`
	response, err := http.Post("http://localhost:32234/NetworkDriver.ProgramExternalConnectivity",
		sdk.DefaultContentTypeV1_1,
		strings.NewReader(request))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("Expected %d, got %d: %s\n", http.StatusOK, response.StatusCode, string(body))
	}
}
