package handlers

import (
	"fmt"
	"net/http"

	ctxu "github.com/docker/distribution/context"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/storage"
	"golang.org/x/net/context"
)

// Context should contain the request specific context for use in across
// handlers. Resources that don't need to be shared across handlers should not
// be on this object.
type Context struct {
	// App points to the application structure that created this context.
	*App
	context.Context

	// Repository is the repository for the current request. All requests
	// should be scoped to a single repository. This field may be nil.
	Repository storage.Repository

	// Errors is a collection of errors encountered during the request to be
	// returned to the client API. If errors are added to the collection, the
	// handler *must not* start the response via http.ResponseWriter.
	Errors v2.Errors

	urlBuilder *v2.URLBuilder

	// TODO(stevvooe): The goal is too completely factor this context and
	// dispatching out of the web application. Ideally, we should lean on
	// context.Context for injection of these resources.
}

// Value overrides context.Context.Value to ensure that calls are routed to
// correct context.
func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.Context.Value(key)
}

func getName(ctx context.Context) (name string) {
	return ctxu.GetStringValue(ctx, "vars.name")
}

func getTag(ctx context.Context) (tag string) {
	return ctxu.GetStringValue(ctx, "vars.tag")
}

var errDigestNotAvailable = fmt.Errorf("digest not available in context")

func getDigest(ctx context.Context) (dgst digest.Digest, err error) {
	dgstStr := ctxu.GetStringValue(ctx, "vars.digest")

	if dgstStr == "" {
		ctxu.GetLogger(ctx).Errorf("digest not available")
		return "", errDigestNotAvailable
	}

	d, err := digest.ParseDigest(dgstStr)
	if err != nil {
		ctxu.GetLogger(ctx).Errorf("error parsing digest=%q: %v", dgstStr, err)
		return "", err
	}

	return d, nil
}

func getUploadUUID(ctx context.Context) (uuid string) {
	return ctxu.GetStringValue(ctx, "vars.uuid")
}

// getUserName attempts to resolve a username from the context and request. If
// a username cannot be resolved, the empty string is returned.
func getUserName(ctx context.Context, r *http.Request) string {
	username := ctxu.GetStringValue(ctx, "auth.user.name")

	// Fallback to request user with basic auth
	if username == "" {
		var ok bool
		uname, _, ok := basicAuth(r)
		if ok {
			username = uname
		}
	}

	return username
}
