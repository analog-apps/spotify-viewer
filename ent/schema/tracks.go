package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/albe2669/spotify-viewer/ent/schema/pulid"
)

type Track struct {
	ent.Schema
}

func (Track) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AtTimeMixin{},
		pulid.MixinWithPrefix("TR"),
	}
}

func (Track) Fields() []ent.Field {
	return []ent.Field{
		field.String("track_id").NotEmpty().GoType(pulid.ID("")),
		field.String("name"),
		field.Strings("artists"),
		field.Strings("artists_genres"),
		field.String("album_name"),
		field.String("album_image_uri"),
		field.Int32("duration_ms"),
		field.String("uri"),
	}
}
