// Package srt contains the SRT static source.
package srt

import (
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/description"
	mcmpegts "github.com/bluenviron/mediacommon/pkg/formats/mpegts"
	"github.com/datarhei/gosrt"

	"github.com/devsolux/mediaserver/internal/conf"
	"github.com/devsolux/mediaserver/internal/defs"
	"github.com/devsolux/mediaserver/internal/logger"
	"github.com/devsolux/mediaserver/internal/protocols/mpegts"
	"github.com/devsolux/mediaserver/internal/stream"
)

// Source is a SRT static source.
type Source struct {
	ReadTimeout conf.StringDuration
	Parent      defs.StaticSourceParent
}

// Log implements StaticSource.
func (s *Source) Log(level logger.Level, format string, args ...interface{}) {
	s.Parent.Log(level, "[SRT source] "+format, args...)
}

// Run implements StaticSource.
func (s *Source) Run(params defs.StaticSourceRunParams) error {
	s.Log(logger.Debug, "connecting")

	conf := srt.DefaultConfig()
	address, err := conf.UnmarshalURL(params.Conf.Source)
	if err != nil {
		return err
	}

	err = conf.Validate()
	if err != nil {
		return err
	}

	sconn, err := srt.Dial("srt", address, conf)
	if err != nil {
		return err
	}

	readDone := make(chan error)
	go func() {
		readDone <- s.runReader(sconn)
	}()

	for {
		select {
		case err := <-readDone:
			sconn.Close()
			return err

		case <-params.ReloadConf:

		case <-params.Context.Done():
			sconn.Close()
			<-readDone
			return nil
		}
	}
}

func (s *Source) runReader(sconn srt.Conn) error {
	sconn.SetReadDeadline(time.Now().Add(time.Duration(s.ReadTimeout)))
	r, err := mcmpegts.NewReader(mcmpegts.NewBufferedReader(sconn))
	if err != nil {
		return err
	}

	decodeErrLogger := logger.NewLimitedLogger(s)

	r.OnDecodeError(func(err error) {
		decodeErrLogger.Log(logger.Warn, err.Error())
	})

	var stream *stream.Stream

	medias, err := mpegts.ToStream(r, &stream)
	if err != nil {
		return err
	}

	res := s.Parent.SetReady(defs.PathSourceStaticSetReadyReq{
		Desc:               &description.Session{Medias: medias},
		GenerateRTPPackets: true,
	})
	if res.Err != nil {
		return res.Err
	}

	stream = res.Stream

	for {
		sconn.SetReadDeadline(time.Now().Add(time.Duration(s.ReadTimeout)))
		err := r.Read()
		if err != nil {
			return err
		}
	}
}

// APISourceDescribe implements StaticSource.
func (*Source) APISourceDescribe() defs.APIPathSourceOrReader {
	return defs.APIPathSourceOrReader{
		Type: "srtSource",
		ID:   "",
	}
}
