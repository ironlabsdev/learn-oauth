package pages

import (
	"net/http"

	"oauth/views/pages"
)

func (p *Pages) Login(w http.ResponseWriter, r *http.Request) {
	err := pages.Login().Render(r.Context(), w)
	if err != nil {
		p.logger.Err(err).Msg("Error occurred in rendering login page")
		return
	}
}
