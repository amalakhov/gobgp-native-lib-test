package bgp

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/uuid"
	api "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
	"log"
	"os"
	"strings"
	"time"
)

type UpdateType int

const (
	ANNOUNCE UpdateType = 0
	WITHDRAW UpdateType = 1
)

type SessionState string

const (
	UNKNOWN     SessionState = "UNKNOWN"
	IDLE        SessionState = "IDLE"
	CONNECT     SessionState = "CONNECT"
	ACTIVE      SessionState = "ACTIVE"
	OPENSENT    SessionState = "OPENSENT"
	OPENCONFIRM SessionState = "OPENCONFIRM"
	ESTABLISHED SessionState = "ESTABLISHED"
)

type PeerStateChangedNotification struct {
	Peer *api.Peer
}

type PeerState struct {
	SessionState SessionState `json:"sessionState"`
	Ts           int64        `json:"ts"`
}

type Neighbor struct {
	NeighborAddress string
	PeerAs          uint32
	LocalAs         uint32
}

func (n *Neighbor) getPeer() (peer *api.Peer) {
	peer = &api.Peer{

		Conf: &api.PeerConf{
			NeighborAddress: n.NeighborAddress,
			PeerAs:          n.PeerAs,
			LocalAs:         n.LocalAs,
		},
	}

	return peer
}

type Configuration struct {
	As       uint32
	RouterId string
	Neighbor Neighbor
}

type UpdateMessage struct {
	Prefix     string     `json:"prefix"`
	NextHop    string     `json:"nextHop"`
	UpdateType UpdateType `json:"updateType"`
	AsPath     string     `json:"asPath"`
	Uuid       string     `json:"uuid"`
}

type MessageHandler struct {
	server        *gobgp.BgpServer
	configuration *Configuration
	States        map[string]PeerState
	In            chan UpdateMessage
	PeerStateOut  chan PeerStateChangedNotification
}

func New(configuration *Configuration) (mh *MessageHandler) {

	states := make(map[string]PeerState)
	states[configuration.Neighbor.NeighborAddress] = PeerState{}

	mh = &MessageHandler{
		server:        gobgp.NewBgpServer(),
		configuration: configuration,
		In:            make(chan UpdateMessage),
		PeerStateOut:  make(chan PeerStateChangedNotification),
		States:        states,
	}

	return mh
}

func (mh *MessageHandler) Run() {
	go mh.server.Serve()

	mh.startBgpServer()
	mh.startPeerStateMonitoring()
	//mh.addPeer(&mh.configuration.Neighbor)

	go mh.runMessageListener()

	go func() {
		for {

			for fo := 1; fo < 255; fo++ {
				_, err := mh.announce(&UpdateMessage{
					Prefix:     fmt.Sprintf("%d.0.128.0/24", fo),
					NextHop:    "172.31.225.46",
					UpdateType: ANNOUNCE,
					AsPath:     "46786 38040 23969",
				})

				if err != nil {
					fmt.Printf("%v\r\n", err)
				}
			}

			for fo := 1; fo < 255; fo++ {
				err := mh.withdraw(&UpdateMessage{
					Prefix:     fmt.Sprintf("%d.0.128.0/24", fo),
					NextHop:    "172.31.225.46",
					UpdateType: WITHDRAW,
					AsPath:     "46786 38040 23969",
				})

				if err != nil {
					fmt.Printf("%v\r\n", err)
				}
			}
		}
	}()
}

func (mh *MessageHandler) runMessageListener() {
	for {
		select {

		case peerState := <-mh.PeerStateOut:
			{
				mh.applyNeighborState(&peerState)
			}

		case updateMsg := <-mh.In:
			{
				if _, err := mh.handle(&updateMsg); err != nil {
					log.Printf("handle: message %v was handled with error %v\n", updateMsg, err)
				}
			}

		}
	}
}

func (mh *MessageHandler) startBgpServer() {
	err := mh.server.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			As:       mh.configuration.As,
			RouterId: mh.configuration.RouterId,
		},
	})

	if err != nil {
		log.Println(err)
		os.Exit(-2)
	}

	log.Println("BgpServer started...")
}

func (mh *MessageHandler) startPeerStateMonitoring() {
	if err := mh.server.MonitorPeer(context.Background(), &api.MonitorPeerRequest{}, func(p *api.Peer) {
		mh.PeerStateOut <- PeerStateChangedNotification{Peer: p}
	}); err != nil {
		log.Println(err)
		os.Exit(-2)
	}

	log.Println("Peer state monitoring started...")
}

func (mh *MessageHandler) addPeer(neighbor *Neighbor) {
	peer := neighbor.getPeer()

	if err := mh.server.AddPeer(context.Background(), &api.AddPeerRequest{
		Peer: peer,
	}); err != nil {
		log.Println(err)
		os.Exit(-2)
	}
}

func (mh *MessageHandler) GetPaths(tableType api.TableType) []*api.Destination {
	v4Family := &api.Family{
		Afi:  api.Family_AFI_IP,
		Safi: api.Family_SAFI_UNICAST,
	}

	var destinations []*api.Destination
	err := mh.server.ListPath(context.Background(), &api.ListPathRequest{
		TableType: tableType,
		Family:    v4Family,
	}, func(p *api.Destination) {
		destinations = append(destinations, p)
	})

	if err != nil {
		log.Println(err)
		return nil
	} else {
		return destinations
	}
}

func (mh *MessageHandler) Handle(msg *UpdateMessage) (string, error) {
	switch msg.UpdateType {

	case ANNOUNCE:
		id, err := mh.announce(msg)
		return id, err

	case WITHDRAW:
		err := mh.withdraw(msg)
		return "", err

	default:
		return "", fmt.Errorf("handle: update type %d isn't supported", msg.UpdateType)
	}
}

func (mh *MessageHandler) applyNeighborState(newState *PeerStateChangedNotification) {
	ts := time.Now().Unix()

	if _, ok := mh.States[newState.Peer.State.NeighborAddress]; ok {
		mh.States[newState.Peer.State.NeighborAddress] = PeerState{
			SessionState: SessionState(newState.Peer.State.SessionState.String()),
			Ts:           ts,
		}

		log.Printf("applyNeighborState: Neighbor state changed. Current states %v\n", mh.States)
	}
}

func (mh *MessageHandler) handle(msg *UpdateMessage) (string, error) {
	switch msg.UpdateType {

	case ANNOUNCE:
		id, err := mh.announce(msg)
		return id, err
	case WITHDRAW:
		err := mh.withdraw(msg)
		return "", err
	default:
		return "", fmt.Errorf("handle: update type %d isn't supported", msg.UpdateType)
	}
}

func (mh *MessageHandler) announce(msg *UpdateMessage) (string, error) {
	prefixParts := strings.Split(msg.Prefix, "/")
	if len(prefixParts) != 2 {
		return "", errors.New("announce: wrong prefix format")
	}

	numbers := stringsToNumbers(strings.Split(msg.AsPath, " "))

	nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
		Prefix:    prefixParts[0],
		PrefixLen: toUint32(prefixParts[1]),
	})

	originAttribute, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 0,
	})

	nextHopAttribute, _ := ptypes.MarshalAny(&api.NextHopAttribute{
		NextHop: msg.NextHop,
	})

	asPathAttribute, _ := ptypes.MarshalAny(&api.AsPathAttribute{
		Segments: []*api.AsSegment{
			{
				Type:    2,
				Numbers: numbers,
			},
		},
	})

	attrs := []*any.Any{originAttribute, nextHopAttribute, asPathAttribute}

	response, err := mh.server.AddPath(context.Background(), &api.AddPathRequest{
		TableType: api.TableType_GLOBAL,

		Path: &api.Path{
			Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})

	if response != nil {

		if id, err := uuid.FromBytes(response.Uuid); err == nil {
			return id.String(), nil
		}

	}

	return "", err
}

func (mh *MessageHandler) withdraw(msg *UpdateMessage) error {
	prefixParts := strings.Split(msg.Prefix, "/")
	if len(prefixParts) != 2 {
		return errors.New("withdraw: wrong prefix format")
	}

	nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
		Prefix:    prefixParts[0],
		PrefixLen: toUint32(prefixParts[1]),
	})

	originAttribute, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 0,
	})

	nextHopAttribute, _ := ptypes.MarshalAny(&api.NextHopAttribute{
		NextHop: msg.NextHop,
	})

	attrs := []*any.Any{originAttribute, nextHopAttribute}

	err := mh.server.DeletePath(context.Background(), &api.DeletePathRequest{
		TableType: api.TableType_GLOBAL,

		Path: &api.Path{
			Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})

	return err
}
