package main

import (
	"context"
	"os"
	"time"

	"github.com/emiago/diago"
	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure logging
	lev, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil || lev == zerolog.NoLevel {
		lev = zerolog.InfoLevel
	}
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.StampMicro,
	}).With().Timestamp().Logger().Level(lev)

	// Create SIP user agent
	ua, err := sipgo.NewUA()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create SIP user agent")
	}

	// Configure transport
	transport := diago.Transport{
		Transport:      "udp",
		BindHost:       "100.73.40.116",
		BindPort:       5081,
		RewriteContact: true,
		ExternalHost:   "100.73.40.116",
		ExternalPort:   5081,
	}

	// Create Diago instance
	dg := diago.NewDiago(ua,
		diago.WithTransport(transport),
	)

	// Create context for the call with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Definir host local para os cabeçalhos
	localHost := "100.73.40.116"

	// Configure caller information
	displayName := "Empresa XYZ"
	callerID := "4833800000"

	// Configure SIP headers with custom From header
	headers := []sip.Header{
		// P-Asserted-Identity header
		sip.NewHeader("P-Asserted-Identity", "<sip:"+callerID+"@100.73.40.116>"),

		// Custom From header with proper host
		&sip.FromHeader{
			DisplayName: displayName,
			Address: sip.Uri{
				User: callerID,
				Host: localHost, // Use o mesmo host do seu transporte
				Port: transport.BindPort,
			},
			Params: sip.NewParams(),
		},
	}

	// Set up recipient URI
	recipient := sip.Uri{
		User:    "08821670000",
		Host:    "100.81.118.20",
		Port:    5080,
		Headers: sip.NewParams(),
	}

	// Configure invite options
	opts := diago.InviteOptions{
		Headers: headers,
		//Media:   diago.DefaultMediaConfig(),
		OnResponse: func(resp *sip.Response) error {
			if resp.StatusCode == 200 {
				log.Info().
					Int("status", resp.StatusCode).
					Str("reason", resp.Reason).
					Msg("Chamada atendida")

				// Marque que recebeu 200 OK
			}
			return nil
		},
	}

	// Place the call
	log.Info().Msg("Starting SIP call...")
	dialog, err := dg.Invite(ctx, recipient, opts)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create dialog")
	}

	defer dialog.Close()
	defer dialog.Bye(ctx)

	// Aguarde um tempo para manter a chamada ativa
	log.Info().Msg("Mantendo chamada ativa...")
	time.Sleep(30 * time.Second)

	log.Info().Msg("Call terminated")
}
