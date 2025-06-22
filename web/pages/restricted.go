package pages

import (
	"net/http"
	"oauth/web/requests/auth"
	"oauth/views/pages"
)

func (p *Pages) Restricted(w http.ResponseWriter, r *http.Request) {
	session, _ := p.store.Get(r, auth.SessionID)
	isAuthenticated := session.Values[auth.AuthenticatedKey]

	if isAuthenticated != nil && isAuthenticated == true {
		// User is authenticated - show restricted content
		userID := session.Values[auth.UserIDKey].(string)
		component := pages.Restricted(userID, "Welcome to the secret area!")
		component.Render(r.Context(), w)
	} else {
		// User is not authenticated - show unauthorized message
		component := pages.Unauthorized()
		component.Render(r.Context(), w)
	}

}
