package api

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/oksa-sy/WASAText/service/api/reqcontext"
	"github.com/sirupsen/logrus"
)

// httpRouterHandler is the signature for functions that accepts a reqcontext.RequestContext in addition to those
// required by the httprouter package.
type httpRouterHandler func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)

// wrap parses the request and adds a reqcontext.RequestContext instance related to the request.
// crea un id per ogni richiesta e lo memorizza insieme l'indirizzo ip del richiedente
func (rt *_router) wrap(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var ctx = reqcontext.RequestContext{
			ReqUUID: reqUUID,
		}

		// Create a request-specific logger
		ctx.Logger = rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     ctx.ReqUUID.String(),
			"remote-ip": r.RemoteAddr,
		})

		// controllo se nell'header è presente 'Bearer'
		header := r.Header.Get("Authorization")
		if len(header) < 7 || header[:7] != "Bearer " {
			ctx.Logger.Error("unauthorized request: missing or invalid token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// controllo se l'id associato è un id valido presente nel database
		ctx.UserId = header[7:]
		exists, err := rt.db.SearchUser(ctx.UserId)
		if err != nil {
			ctx.Logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !exists {
			ctx.Logger.Error("unauthorized request: missing or invalid id")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Call the next handler in chain (usually, the handler function for the path)
		fn(w, r, ps, ctx)
	}
}
