package dockerutil

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/docker/go-connections/nat"
)

var mu sync.RWMutex

type Listeners []net.Listener

func (l Listeners) CloseAll() {
	for _, listener := range l {
		listener.Close()
	}
}

// openListener opens a listener on a port. Set to 0 to get a random port.
func openListener(port int) (*net.TCPListener, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	mu.Lock()
	defer mu.Unlock()
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// getPort generates a docker PortBinding by using the port provided.
// If port is set to 0, the next avaliable port will be used.
// The listener will be closed in the case of an error, otherwise it will be left open.
// This allows multiple getPort calls to find multiple available ports
// before closing them so they are available for the PortBinding.
func getPort(port int) (nat.PortBinding, *net.TCPListener, error) {
	l, err := openListener(port)
	if err != nil {
		l.Close()
		return nat.PortBinding{}, nil, err
	}

	return nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: fmt.Sprint(l.Addr().(*net.TCPAddr).Port),
	}, l, nil
}

// GeneratePortBindings will find open ports on the local
// machine and create a PortBinding for every port in the portSet.
// If a port is already bound, it will use that port as an override.
func GeneratePortBindings(pairs nat.PortMap) (nat.PortMap, Listeners, error) {
	m := make(nat.PortMap)
	listeners := make(Listeners, 0, len(pairs))

	var pb nat.PortBinding
	var l *net.TCPListener
	var err error

	for p, bind := range pairs {
		if len(bind) == 0 {
			// random port
			pb, l, err = getPort(0)
		} else {
			var pNum int
			if pNum, err = strconv.Atoi(bind[0].HostPort); err != nil {
				return nat.PortMap{}, nil, err
			}

			pb, l, err = getPort(pNum)
		}

		if err != nil {
			listeners.CloseAll()
			return nat.PortMap{}, nil, err
		}

		listeners = append(listeners, l)
		m[p] = []nat.PortBinding{pb}
	}

	return m, listeners, nil
}
