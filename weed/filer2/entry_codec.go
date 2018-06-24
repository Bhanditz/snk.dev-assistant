package filer2

import (
	"os"
	"time"

	"fmt"
	"github.com/chrislusf/seaweedfs/weed/pb/filer_pb"
	"github.com/gogo/protobuf/proto"
)

func (entry *Entry) EncodeAttributesAndChunks() ([]byte, error) {
	message := &filer_pb.Entry{
		Attributes: EntryAttributeToPb(entry),
		Chunks:     entry.Chunks,
	}
	return proto.Marshal(message)
}

func (entry *Entry) DecodeAttributesAndChunks(blob []byte) error {

	message := &filer_pb.Entry{}

	if err := proto.UnmarshalMerge(blob, message); err != nil {
		return fmt.Errorf("decoding value blob for %s: %v", entry.FullPath, err)
	}

	entry.Attr = PbToEntryAttribute(message.Attributes)

	entry.Chunks = message.Chunks

	return nil
}

func EntryAttributeToPb(entry *Entry) *filer_pb.FuseAttributes {

	return &filer_pb.FuseAttributes{
		Crtime:      entry.Attr.Crtime.Unix(),
		Mtime:       entry.Attr.Mtime.Unix(),
		FileMode:    uint32(entry.Attr.Mode),
		Uid:         entry.Uid,
		Gid:         entry.Gid,
		Mime:        entry.Mime,
		Collection:  entry.Attr.Collection,
		Replication: entry.Attr.Replication,
		TtlSec:      entry.Attr.TtlSec,
	}
}

func PbToEntryAttribute(attr *filer_pb.FuseAttributes) Attr {

	t := Attr{}

	t.Crtime = time.Unix(attr.Crtime, 0)
	t.Mtime = time.Unix(attr.Mtime, 0)
	t.Mode = os.FileMode(attr.FileMode)
	t.Uid = attr.Uid
	t.Gid = attr.Gid
	t.Mime = attr.Mime
	t.Collection = attr.Collection
	t.Replication = attr.Replication
	t.TtlSec = attr.TtlSec

	return t
}
