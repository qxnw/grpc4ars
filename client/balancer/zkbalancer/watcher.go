package zkbalancer

import (
	"fmt"

	"github.com/qxnw/lib4go/zk"

	"google.golang.org/grpc/naming"
)

// watcher is the implementaion of grpc.naming.Watcher
type watcher struct {
	re            *resolver // re: Etcd Resolver
	client        *zk.ZookeeperClient
	isInitialized bool
	caches        map[string]bool
}

// Close do nothing
func (w *watcher) Close() {
}

// Next to return the updates
func (w *watcher) Next() ([]*naming.Update, error) {
	// prefix is the etcd prefix/value to watch
	servicePath := fmt.Sprintf("%s", w.re.serviceRoot)
	updates := make([]*naming.Update, 0, 4)
	// check if is initialized
	if !w.isInitialized {
		// query addresses from etcd
		servicePath := fmt.Sprintf("%s", w.re.serviceRoot)
		resp, err := w.client.GetChildren(servicePath)
		w.isInitialized = true
		if err == nil {
			addrs := extractAddrs(resp)
			//if not empty, return the updates or watcher new dir
			if l := len(addrs); l != 0 {
				for i := range addrs {
					updates = append(updates, &naming.Update{Op: naming.Add, Addr: addrs[i]})
				}
				return updates, nil
			}
		}
	}

	// generate etcd Watcher
	resp, err := w.client.WatchChildren(servicePath)
	if err != nil {
		return nil, err
	}
	addrs := extractAddrs(resp)
	newCache := make(map[string]bool)
	for i := 0; i < len(addrs); i++ {
		if _, ok := w.caches[addrs[i]]; !ok {
			updates = append(updates, &naming.Update{Op: naming.Add, Addr: addrs[i]})
			newCache[addrs[i]] = true
		} else {
			w.caches[addrs[i]] = false
		}
	}
	for i, v := range w.caches {
		if v {
			updates = append(updates, &naming.Update{Op: naming.Delete, Addr: i})
		}
	}
	w.caches = newCache
	return updates, nil
}

func extractAddrs(resp []string) []string {
	return resp
}
