package main

import (
	"admission-webhook-pod/k8s"
	"admission-webhook-pod/options"
	"admission-webhook-pod/webhook"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func main() {
	parameters := options.OptionsParams{}
	parameters.FlagParse()

	level := log.Level(parameters.LogLevel)
	log.SetLevel(level)
	log.Printf("Log level: %s", log.GetLevel())
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	k := k8s.NewK8S(&parameters)
	if err := k.Run(); err != nil {
		log.Panic(err)
	}

	pair, err := tls.X509KeyPair(k.CertPEM.Bytes(), k.CertKeyPEM.Bytes())
	if err != nil {
		log.Errorf("Failed to load key pair: %v", pair)
	}

	whsvr := &webhook.WebhookServer{
		Server: &http.Server{
			Addr: fmt.Sprintf(":%v", parameters.Port),
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{pair},
			},
		},
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.Handle(options.MutatePath, http.HandlerFunc(whsvr.Serve))
	whsvr.Server.Handler = mux

	// start webhook server in new rountine
	go func() {
		if err := whsvr.Server.ListenAndServeTLS("", ""); err != nil {
			log.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	log.Infof("Server started, listening to the port %d", parameters.Port)

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	whsvr.Server.Shutdown(context.Background())
}
