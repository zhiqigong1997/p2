package storageserver

import (
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
	"strconv"

	"github.com/cmu440/tribbler/rpc/storagerpc"
	"time"
)

type role int
type status int

type storageServer struct {
	// TODO: implement this!
	isMaster    bool
	mux         sync.Mutex
	serverList  map[string]*storageServer
	masterHost  string
	masterPort  int
	numNodes    int
	listenPort  int
	nodeID      uint32
	activeNodes int
	joinNode    chan storagerpc.Node
	// userMap  map[string]
}

// NewStorageServer creates and starts a new StorageServer. masterServerHostPort
// is the master storage server's host:port address. If empty, then this server
// is the master; otherwise, this server is a slave. numNodes is the total number of
// servers in the ring. port is the port number that this server should listen on.
// nodeID is a random, unsigned 32-bit ID identifying this server.
//
// This function should return only once all storage servers have joined the ring,
// and should return a non-nil error if the storage server could not be started.
func NewStorageServer(masterServerHostPort string, numNodes, port int, nodeID uint32) (StorageServer, error) {
	srv := new(storageServer)
	if masterServerHostPort == "" {
		srv.isMaster = true
		srv.numNodes = numNodes
		srv.masterPort = port
		srv.activeNodes = 0
		srv.joinNode = make(chan storagerpc.Node)

	}
	srv.listenPort = port
	srv.nodeID = nodeID

	if !srv.isMaster {
		cli, err := rpc.DialHTTP("tcp", masterServerHostPort)
		if err != nil {
			return nil, err
		}
		nodeInfo := storagerpc.Node{NodeID: nodeID, HostPort: fmt.Sprintf(":%d", port)}
		args := storagerpc.RegisterArgs{nodeInfo}
		var reply storagerpc.RegisterReply

		for {
			if err = cli.Call("StorageServer.RegisterServer", args, &reply); err != nil {
				return nil, err
			}
			if reply.Status == storagerpc.OK {
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}
		return nil, fmt.Errorf("Returned status %d", reply.Status)
	} else { // if master, then listen to RPC calls, until all joined

		if err := rpc.Register(&srv); err != nil {
			return nil, err
		}
		rpc.HandleHTTP()
		go func() {
			err := http.ListenAndServe(":"+strconv.Itoa(srv.listenPort), nil)
			if err != nil {
				fmt.Errorf("http server start failed with error:", err)
			}
		}()

		select {
		case <-srv.joinNode:
			srv.activeNodes++
			if srv.activeNodes == srv.numNodes {
				break
			}
		}
		return srv, nil
	}
}

func (ss *storageServer) RegisterServer(args *storagerpc.RegisterArgs, reply *storagerpc.RegisterReply) error {
	ss.joinNode <- args.ServerInfo
	ss.mux.RLock()
	if ss.activeNodes < ss.numNodes {
		reply.Status = storagerpc.NotReady
	} else {
		reply.Status = storagerpc.OK
	}
	reply.Servers = make([]storagerpc.Node, len(ss.activeNodes))
	copy(reply.Servers, ss.activeNodes)

	ss.mux.UnRlock()

	return nil
}

func (ss *storageServer) GetServers(args *storagerpc.GetServersArgs, reply *storagerpc.GetServersReply) error {
	ss.mux.RLock()
	if ss.activeNodes < ss.numNodes {
		reply.Status = storagerpc.NotReady
	} else {
		reply.Status = storagerpc.OK
	}
	reply.Servers = make([]storagerpc.Node, len(ss.activeNodes))
	copy(reply.Servers, ss.activeNodes)
	ss.mux.UnRlock()

	return errors.New("not implemented")
}

func (ss *storageServer) Get(args *storagerpc.GetArgs, reply *storagerpc.GetReply) error {
	return errors.New("not implemented")
}

func (ss *storageServer) Delete(args *storagerpc.DeleteArgs, reply *storagerpc.DeleteReply) error {
	return errors.New("not implemented")
}

func (ss *storageServer) GetList(args *storagerpc.GetArgs, reply *storagerpc.GetListReply) error {
	return errors.New("not implemented")
}

func (ss *storageServer) Put(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	return errors.New("not implemented")
}

func (ss *storageServer) AppendToList(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	return errors.New("not implemented")
}

func (ss *storageServer) RemoveFromList(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	return errors.New("not implemented")
}
