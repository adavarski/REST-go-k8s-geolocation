package controllers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/adavarski/REST-go-k8s-geolocation/internal/models"
	"github.com/rs/zerolog/hlog"
)

const (
	// ContentTypeApplicationJSON represent the applcation/json Content-Type value
	ContentTypeApplicationJSON = "application/json"
)

// GetGeoIP is the main handler.
// It will parse the route variable to ensure it is a valid IPv4 address
// before getting the GeoIP information for the given address.
// It will take care of updating the caches if necessary.
func (h *BaseHandler) GetGeoIP(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	req_id, _ := hlog.IDFromCtx(r.Context())

	// Get ip from URL and parse it to a net.IP
	params := httprouter.ParamsFromContext(r.Context())
	ip := params.ByName("ip")
	if !isIpv4(ip) {
		h.Logger.Error().Str("req_id", req_id.String()).Msg("the provided IP is not a valid IPv4 address")
		e := NewErrorResponse("the provided IP is not a valid IPv4 address")
		resp, _ := json.Marshal(e)
		w.Header().Set("Content-Type", ContentTypeApplicationJSON)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	var req_ctx = r.Context()
	var ctx = context.Background()
	var g *models.GeoIP
	var err error

	// WaitGroup for cache update goroutines
	// wg.Wait() will not be called as the Save() operation
	// can be executed in the background after the http request has
	// been finished.
	var wg sync.WaitGroup

	// TODO: Implement custom errors

	// Retrieve the IP information from the in-memory database
	g, err = h.InMemoryRepo.Get(ctx, ip)
	if err != nil {
		h.Logger.Debug().Str("req_id", req_id.String()).Msg("cache miss from in-memory database")
		// Retrieve the IP information from the redis database
		g, err = h.RedisRepo.Get(ctx, ip)
		if err != nil {
			h.Logger.Debug().Str("req_id", req_id.String()).Msg("cache miss from redis database")
			// Query the remote GeoIP API to retrieve IP information
			g, err = h.RemoteIPAPI.Get(req_ctx, ip)
			if err != nil {
				h.Logger.Debug().Str("req_id", req_id.String()).Msg("couldn't retrieve geo IP information")
				e := NewErrorResponse("Couldn't retrieve geo IP information")
				resp, _ := json.Marshal(e)
				w.Header().Set("Content-Type", ContentTypeApplicationJSON)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(resp)
				return
			}

			// Save the IP information in the redis and the in-memory databases
			// for later use
			wg.Add(2)

			// Populate the in-memory local cache with the new value
			go func() {
				defer wg.Done()

				h.Logger.Debug().Str("req_id", req_id.String()).Msg("caching in in-memory database")
				if err := h.InMemoryRepo.Save(ctx, g); err != nil {
					h.Logger.Error().Str("req_id", req_id.String()).Msg("fail to cache in in-memory database: " + err.Error())
				}
			}()

			// Populate the Redis cache with the new value
			go func() {
				defer wg.Done()

				h.Logger.Debug().Str("req_id", req_id.String()).Msg("caching in redis database")
				if err := h.RedisRepo.Save(ctx, g); err != nil {
					h.Logger.Error().Str("req_id", req_id.String()).Msg("fail to cache in redis database: " + err.Error())
				}
			}()
		} else {
			h.Logger.Debug().Str("req_id", req_id.String()).Msg("cache hit from redis database")

			// Save the IP information in the in-memory databases
			// for later use
			wg.Add(1)

			// Populate the in-memory local cache with the new value
			go func() {
				defer wg.Done()

				h.Logger.Debug().Str("req_id", req_id.String()).Msg("caching in in-memory database")
				if err := h.InMemoryRepo.Save(ctx, g); err != nil {
					h.Logger.Error().Str("req_id", req_id.String()).Msg("fail to cache in in-memory database: " + err.Error())
				}
			}()
		}
	} else {
		h.Logger.Debug().Str("req_id", req_id.String()).Msg("cache hit from in-memory database")

		// Verify whether the key is in redis, if no, add it
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Retrieve the IP information from the redis database
			_, err := h.RedisRepo.Get(ctx, ip)
			if err != nil {
				h.Logger.Debug().Str("req_id", req_id.String()).Msg("cache miss from redis database... Adding to cache")
				if err := h.RedisRepo.Save(ctx, g); err != nil {
					h.Logger.Error().Str("req_id", req_id.String()).Msg("fail to cache in redis database: " + err.Error())
				}
			}
		}()
	}

	// Marshal the response in json format
	resp, err := json.Marshal(g)
	if err != nil {
		h.Logger.Error().Str("req_id", req_id.String()).Msg("couldn't marshal geo IP information")
		e := NewErrorResponse("couldn't marshal geo IP information")
		resp, _ := json.Marshal(e)
		w.Header().Set("Content-Type", ContentTypeApplicationJSON)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp)
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// isIpv4 verify the given string is a valid IPv4 address.
// Return true if yes, false otherwise
func isIpv4(host string) bool {
	return net.ParseIP(host) != nil
}
