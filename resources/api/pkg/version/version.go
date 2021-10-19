package version

import (
	"net/http"

	"github.com/go-chi/render"
)

// Version of the program.
var Version = "SNAPSHOT"

// Commit Hash
var BuildHash = "AAAAAAAA"

// date the program was built
var BuildDate = "19760101"

//
//
//

type versionResponse struct {
	Version string `json:"version"`

	BuildHash string `json:"buildhash"`

	BuildDate string `json:"builddate"`
}

func (u *versionResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Handler used to return buildversion
func Handler(w http.ResponseWriter, r *http.Request) {

	render.Status(r, http.StatusOK)

	render.Render(w, r, &versionResponse{
		Version:   Version,
		BuildHash: BuildHash,
		BuildDate: BuildDate,
	})
}
