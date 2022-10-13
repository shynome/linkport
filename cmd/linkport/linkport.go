package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/donovanhide/eventsource"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/pion/webrtc/v3"
	"github.com/shynome/wl"
	"github.com/shynome/wl/ortc"
)

var args struct {
	endpoint   string
	isServer   bool
	port       string
	user, pass string
}

func init() {
	flag.BoolVar(&args.isServer, "server", isServer, "")
	flag.StringVar(&args.endpoint, "api", "https://lens.slive.fun/", "")
	flag.StringVar(&args.user, "user", "test", "")
	flag.StringVar(&args.pass, "pass", "vvv", "")
	flag.StringVar(&args.port, "port", "", "")
}

func main() {
	flag.Parse()

	if args.user == "" {
		fmt.Println("flag api user info is required")
		return
	}

	if args.isServer {
		runServer()
		return
	}

	runClient()
}

var wrtcApi = func() (api *webrtc.API) {

	settingEngine := webrtc.SettingEngine{}
	settingEngine.DetachDataChannels()

	api = webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))

	return
}()

const topic = "wrtc-linkport"

func runServer() {

	if args.user == "" || args.pass == "" {
		fmt.Println("flag api user and pass is required")
		return
	}

	finishTask := func(topic string, id string, input []byte) (err error) {
		defer err2.Return(&err)
		link := try.To1(WithTopic(args.endpoint, topic))
		req := try.To1(http.NewRequest(http.MethodDelete, link, bytes.NewReader(input)))
		req.SetBasicAuth(args.user, args.pass)
		req.Header.Set("X-Event-Id", id)
		resp := try.To1(http.DefaultClient.Do(req))
		try.To(CheckResp(resp))
		return
	}
	l := wl.Listen()
	link := try.To1(WithTopic(args.endpoint, topic))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello world")
	})
	go http.Serve(l, nil)

	var req = try.To1(http.NewRequest(http.MethodGet, link, nil))
	req.SetBasicAuth(args.user, args.pass)
	stream := try.To1(eventsource.SubscribeWithRequest("", req))
	fmt.Println("worker start")
	for ev := range stream.Events {
		go func(ev eventsource.Event) {
			var err error
			defer func() {
				if err != nil {
					fmt.Println("server connect err:", err)
				}
			}()
			defer err2.Return(&err)
			fmt.Println("msg event:", ev.Event())
			var offer ortc.Signal
			try.To(json.Unmarshal([]byte(ev.Data()), &offer))
			var pc = try.To1(wrtcApi.NewPeerConnection(webrtc.Configuration{}))
			roffer := try.To1(ortc.HandleConnect(pc, offer))
			rofferBytes := try.To1(json.Marshal(roffer))
			try.To(finishTask(topic, ev.Id(), rofferBytes))
			peer := &wl.Peer{PC: pc}
			l.Add(peer)
			pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
				if pcs == webrtc.PeerConnectionStateDisconnected {
					fmt.Println("remove peer")
					l.Remove(peer)
					peer.Close()
				}
			})
		}(ev)
	}
}

func runClient() {
	if args.user == "" {
		fmt.Println("flag api user is required")
		return
	}

	if args.port == "" {
		fmt.Println("flag port is required")
		return
	}

	call := func(topic string, input []byte) (output []byte, err error) {
		defer err2.Return(&err)

		endpoint := try.To1(WithTopic(args.endpoint, topic))
		req := try.To1(http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(input)))
		req.SetBasicAuth(args.user, "")
		resp := try.To1(http.DefaultClient.Do(req))
		try.To(CheckResp(resp))
		output = try.To1(io.ReadAll(resp.Body))

		return
	}

	pc := try.To1(wrtcApi.NewPeerConnection(webrtc.Configuration{}))
	offer := try.To1(ortc.CreateOffer(pc))
	offerBytes := try.To1(json.Marshal(offer))
	rofferBytes := try.To1(call(topic, offerBytes))
	// fmt.Println("roffer:", string(rofferBytes))
	var roffer ortc.Signal
	try.To(json.Unmarshal(rofferBytes, &roffer))
	try.To(ortc.Handshake(pc, roffer))

	const hostName = "linkport"
	t := wl.NewTransport()
	session := try.To1(wl.NewClientSession(pc))
	t.Set(hostName, session)

	pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		fmt.Println("state2", pcs)
		os.Exit(0)
	})

	l := try.To1(net.Listen("tcp", args.port))
	fmt.Println("forward port start")
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go func(conn net.Conn) {
			defer conn.Close()

			fconn := try.To1(t.NewConn(hostName))
			defer fconn.Close()

			go io.Copy(fconn, conn)
			io.Copy(conn, fconn)
		}(conn)
	}
}
