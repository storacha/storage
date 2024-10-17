package assert

import (
	"net/url"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/validator"
	adm "github.com/storacha/storage/pkg/capability/assert/datamodel"
)

type Range struct {
	Offset uint64
	Length *uint64
}

type LocationCaveats struct {
	Content  multihash.Multihash
	Location []url.URL
	Range    *Range
	Space    did.DID
}

func (lc LocationCaveats) ToIPLD() (datamodel.Node, error) {
	asStrings := make([]string, 0, len(lc.Location))
	for _, location := range lc.Location {
		asStrings = append(asStrings, location.String())
	}

	model := &adm.LocationCaveatsModel{
		Content:  lc.Content,
		Location: asStrings,
		Space:    lc.Space.Bytes(),
	}
	if lc.Range != nil {
		model.Range = &adm.RangeModel{
			Offset: lc.Range.Offset,
			Length: lc.Range.Length,
		}
	}
	return ipld.WrapWithRecovery(model, adm.LocationCaveatsType())
}

const LocationAbility = "assert/location"

var LocationCaveatsReader = schema.Mapped(schema.Struct[adm.LocationCaveatsModel](adm.LocationCaveatsType(), nil), func(model adm.LocationCaveatsModel) (LocationCaveats, failure.Failure) {
	content, err := multihash.Cast(model.Content)
	if err != nil {
		return LocationCaveats{}, failure.FromError(err)
	}

	location := make([]url.URL, 0, len(model.Location))
	for _, l := range model.Location {
		url, err := schema.URI().Read(l)
		if err != nil {
			return LocationCaveats{}, err
		}
		location = append(location, url)
	}

	space := did.Undef
	if len(model.Space) > 0 {
		var serr error
		space, serr = did.Decode(model.Space)
		if serr != nil {
			return LocationCaveats{}, failure.FromError(serr)
		}
	}

	lc := LocationCaveats{
		Content:  content,
		Location: location,
		Space:    space,
	}
	if model.Range != nil {
		lc.Range = &Range{
			Offset: model.Range.Offset,
			Length: model.Range.Length,
		}
	}
	return lc, nil
})

var Location = validator.NewCapability(LocationAbility, schema.DIDString(), LocationCaveatsReader, nil)
