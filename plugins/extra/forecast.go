package extra

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	darksky "github.com/mlbright/darksky/v2"
	"googlemaps.github.io/maps"

	"github.com/belak/go-seabird"
	"github.com/belak/nut"
	"github.com/go-irc/irc"
)

func init() {
	seabird.RegisterPlugin("forecast", newForecastPlugin)
}

type forecastPlugin struct {
	Key        string
	MapsKey    string
	db         *nut.DB
	mapsClient *maps.Client
	// CacheDuration string
}

// ForecastLocation is a simple cache which will store the lat and lon of a
// geocoded location, along with the user who requested this be their home
// location.
type ForecastLocation struct {
	Nick    string
	Address string
	Lat     float64
	Lon     float64
}

func newForecastPlugin(b *seabird.Bot, cm *seabird.CommandMux, db *nut.DB) error {
	p := &forecastPlugin{db: db}

	// Ensure the table is created if it doesn't exist
	err := p.db.EnsureBucket("forecast_location")
	if err != nil {
		return err
	}

	err = b.Config("forecast", p)
	if err != nil {
		return err
	}

	cm.Event("weather", p.weatherCallback, &seabird.HelpInfo{
		Usage:       "<location>",
		Description: "Retrieves current weather for given location",
	})

	cm.Event("forecast", p.forecastCallback, &seabird.HelpInfo{
		Usage:       "<location>",
		Description: "Retrieves three-day forecast for given location",
	})

	options := []maps.ClientOption{}
	if p.MapsKey != "" {
		options = append(options, maps.WithAPIKey(p.MapsKey))
	}

	p.mapsClient, err = maps.NewClient(options...)
	if err != nil {
		return err
	}

	return nil
}

func (p *forecastPlugin) forecastQuery(loc *ForecastLocation) (*darksky.Forecast, error) {
	return darksky.Get(
		p.Key,
		strconv.FormatFloat(loc.Lat, 'f', 4, 64),
		strconv.FormatFloat(loc.Lon, 'f', 4, 64),
		"now",
		darksky.US,
		darksky.English,
	)
}

func (p *forecastPlugin) getLocation(m *irc.Message) (*ForecastLocation, error) {
	l := m.Trailing()

	target := &ForecastLocation{Nick: m.Prefix.Name}

	// If it's an empty string, check the cache
	if l == "" {
		err := p.db.View(func(tx *nut.Tx) error {
			bucket := tx.Bucket("forecast_location")
			return bucket.Get(target.Nick, target)
		})
		if err != nil {
			return nil, fmt.Errorf("Could not find a location for %q", m.Prefix.Name)
		}
		return target, nil
	}

	// If it's not an empty string, we have to look up the location and store
	// it.
	res, err := p.mapsClient.Geocode(context.TODO(), &maps.GeocodingRequest{
		Address: l,
	})
	if err != nil {
		return nil, err
	} else if len(res) == 0 {
		return nil, errors.New("No location results found")
	} else if len(res) > 1 {
		return nil, errors.New("More than 1 result")
	}

	newLocation := &ForecastLocation{
		Nick:    m.Prefix.Name,
		Address: res[0].FormattedAddress,
		Lat:     res[0].Geometry.Location.Lat,
		Lon:     res[0].Geometry.Location.Lng,
	}

	err = p.db.Update(func(tx *nut.Tx) error {
		bucket := tx.Bucket("forecast_location")
		return bucket.Put(newLocation.Nick, newLocation)
	})

	return newLocation, err
}

func (p *forecastPlugin) forecastCallback(b *seabird.Bot, m *irc.Message) {
	loc, err := p.getLocation(m)
	if err != nil {
		b.MentionReply(m, "%s", err.Error())
		return
	}

	fc, err := p.forecastQuery(loc)
	if err != nil {
		b.MentionReply(m, "%s", err.Error())
		return
	}

	b.MentionReply(m, "3 day forecast for %s.", loc.Address)
	for _, block := range fc.Daily.Data[1:4] {
		day := time.Unix(int64(block.Time), 0).Weekday()

		b.MentionReply(m,
			"%s: High %.2f, Low %.2f, Humidity %.f%%. %s",
			day,
			block.TemperatureMax,
			block.TemperatureMin,
			block.Humidity*100,
			block.Summary)
	}
}

func (p *forecastPlugin) weatherCallback(b *seabird.Bot, m *irc.Message) {
	loc, err := p.getLocation(m)
	if err != nil {
		b.MentionReply(m, "%s", err.Error())
		return
	}

	fc, err := p.forecastQuery(loc)
	if err != nil {
		b.MentionReply(m, "%s", err.Error())
		return
	}

	today := fc.Daily.Data[0]
	b.MentionReply(m,
		"%s. Currently %.1f. High %.2f, Low %.2f, Humidity %.f%%. %s.",
		loc.Address,
		fc.Currently.Temperature,
		today.TemperatureMax,
		today.TemperatureMin,
		fc.Currently.Humidity*100,
		fc.Currently.Summary)
}
