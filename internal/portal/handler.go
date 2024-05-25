// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package portal

import "net/http"

func (p *Portal) home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := p.templates.TmplHome.Execute(w, nil)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to execute template", "error", err)
		return
	}
}

func (p *Portal) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := p.templates.TmplLogin.Execute(w, nil)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to execute template", "error", err)
		return
	}
}

func (p *Portal) register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := p.templates.TmplRegister.Execute(w, nil)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to execute template", "error", err)
		return
	}
}
