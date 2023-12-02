package core

import (
	"github.com/devsolux/mediaserver/internal/asyncwriter"
	"github.com/devsolux/mediaserver/internal/defs"
	"github.com/devsolux/mediaserver/internal/stream"
)

// reader is an entity that can read a stream.
type reader interface {
	close()
	apiReaderDescribe() defs.APIPathSourceOrReader
}

func readerMediaInfo(r *asyncwriter.Writer, stream *stream.Stream) string {
	return mediaInfo(stream.MediasForReader(r))
}
