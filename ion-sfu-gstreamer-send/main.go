package main

import (
	"context"
	"flag"
	"fmt"

	gst "github.com/nikunjy/ion-go-examples/pkg/gstreamer-src"

	sdk "github.com/nikunjy/ion-sdk"
	ilog "github.com/pion/ion-log"
	"github.com/pion/webrtc/v3"
)

var (
	log = ilog.NewLoggerWithFields(ilog.DebugLevel, "", nil)
)

func Feed(ctx context.Context, addr string, roomName string) error {
	session := roomName
	codec := "h264"
	videoSrc := "autovideosrc !  videoconvert ! videoscale"
	audioSrc := "autoaudiosrc"
	fmt.Println(videoSrc, audioSrc)
	mimeType := fmt.Sprintf("video/%s", codec)
	// add stun servers
	webrtcCfg := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			webrtc.ICEServer{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	config := sdk.Config{
		WebRTC: sdk.WebRTCTransportConfig{
			Configuration: webrtcCfg,
			VideoMime:     mimeType,
		},
	}

	// new sdk engine
	e := sdk.NewEngine(config)
	// get a client from engine
	c, err := sdk.NewClient(e, addr, "gstreamer")
	if err != nil {
		return err
	}
	var peerConnection *webrtc.PeerConnection = c.GetPubTransport().GetPeerConnection()

	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Infof("Connection state changed: %s", state)
	})

	if err != nil {
		log.Errorf("client err=%v", err)
		return err
	}

	err = e.AddClient(c)
	if err != nil {
		return err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: mimeType, ClockRate: 90000, Channels: 0, SDPFmtpLine: "packetization-mode=1;profile-level-id=42e01f", RTCPFeedback: nil}, "video11", "pion2")
	if err != nil {
		return err
	}

	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		return err
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio11", "pion2")
	if err != nil {
		return err
	}
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		return err
	}

	// client join a session
	err = c.Join(session, sdk.NewJoinConfig().SetNoSubscribe())

	if err != nil {
		log.Errorf("join err=%v", err)
		return err
	}

	// Start pushing buffers on these tracks
	gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{audioTrack}, audioSrc).Start()
	gst.CreatePipeline(codec, []*webrtc.TrackLocalStaticSample{videoTrack}, videoSrc).Start()

	select {
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func main() {
	// parse flag
	var session, addr string

	var codec string
	flag.StringVar(&addr, "addr", "localhost:50051", "Ion-sfu grpc addr")
	flag.StringVar(&session, "session", "test room", "join session name")
	flag.StringVar(&codec, "codec", "vp8", "codec name")

	ctx := context.Background()
	if err := Feed(ctx, addr, session); err != nil {
		panic(err)
	}
	audioSrc := flag.String("audio-src", "audiotestsrc", "GStreamer audio src")
	videoSrc := flag.String("video-src", "videotestsrc", "GStreamer video src")
	flag.Parse()

	if codec != "vp8" && codec != "h264" {
		log.Fatal("No valid codec provided")
	}

	mimeType := fmt.Sprintf("video/%s", codec)
	// add stun servers
	webrtcCfg := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			webrtc.ICEServer{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	config := sdk.Config{
		WebRTC: sdk.WebRTCTransportConfig{
			Configuration: webrtcCfg,
			VideoMime:     mimeType,
		},
	}

	// new sdk engine
	e := sdk.NewEngine(config)
	// get a client from engine
	c, err := sdk.NewClient(e, addr, "gstreamer")

	var peerConnection *webrtc.PeerConnection = c.GetPubTransport().GetPeerConnection()

	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Infof("Connection state changed: %s", state)
	})

	if err != nil {
		log.Errorf("client err=%v", err)
		panic(err)
	}

	err = e.AddClient(c)
	if err != nil {
		return
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: mimeType, ClockRate: 90000, Channels: 0, SDPFmtpLine: "packetization-mode=1;profile-level-id=42e01f", RTCPFeedback: nil}, "video11", "pion2")
	if err != nil {
		panic(err)
	}

	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio11", "pion2")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		panic(err)
	}

	// client join a session
	err = c.Join(session, sdk.NewJoinConfig().SetNoSubscribe())

	if err != nil {
		log.Errorf("join err=%v", err)
		panic(err)
	}

	// Start pushing buffers on these tracks
	gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{audioTrack}, *audioSrc).Start()
	gst.CreatePipeline(codec, []*webrtc.TrackLocalStaticSample{videoTrack}, *videoSrc).Start()

	select {}
}
