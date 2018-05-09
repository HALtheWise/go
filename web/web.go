package web

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/HALtheWise/o-links/context"

	"database/sql"

	_ "github.com/lib/pq"
)

// Serve a bundled asset over HTTP.
func serveAsset(w http.ResponseWriter, r *http.Request, name string) {
	n, err := AssetInfo(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	a, err := Asset(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeContent(w, r, n.Name(), n.ModTime(), bytes.NewReader(a))
}

func templateFromAssetFn(fn func() (*asset, error)) (*template.Template, error) {
	a, err := fn()
	if err != nil {
		return nil, err
	}

	t := template.New(a.info.Name())
	return t.Parse(string(a.bytes))
}

// The default handler responds to most requests. It is responsible for the
// shortcut redirects and for sending unmapped shortcuts to the edit page.
func getDefault(ctx *context.Context, w http.ResponseWriter, r *http.Request) {
	name := parseName("/", r.URL.Path)
	if name == "" {
		http.Redirect(w, r, "/edit/", http.StatusTemporaryRedirect)
		return
	}

	name, err := normalizeName(name)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
	}

	rt, err := ctx.Get(name)
	if err == sql.ErrNoRows {
		http.Redirect(w, r,
			fmt.Sprintf("/edit/%s", cleanName(name)),
			http.StatusTemporaryRedirect)
		return
	} else if err != nil {
		log.Panic(err)
	}

	fmt.Println(fmt.Sprintf("This is the address the database pulled out: %s", rt.URL))

	http.Redirect(w, r,
		rt.URL,
		http.StatusTemporaryRedirect)

}

func getLinks(ctx *context.Context, w http.ResponseWriter, r *http.Request) {
	t, err := templateFromAssetFn(linksHtml)
	if err != nil {
		log.Panic(err)
	}

	rts, err := ctx.GetAll()
	if err != nil {
		log.Panic(err)
	}

	if err := t.Execute(w, rts); err != nil {
		log.Panic(err)
	}
}

// ListenAndServe sets up all web routes, binds the port and handles incoming
// web requests.
func ListenAndServe(addr string, admin bool, version string, ctx *context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/url/", func(w http.ResponseWriter, r *http.Request) {
		apiURL(ctx, w, r)
	})

	/*Taking this one out since it's the most work for the least reward*/
	// mux.HandleFunc("/api/urls/", func(w http.ResponseWriter, r *http.Request) {
	// 	apiURLs(ctx, w, r)
	// })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		getDefault(ctx, w, r)
	})
	mux.HandleFunc("/edit/", func(w http.ResponseWriter, r *http.Request) {
		p := parseName("/edit/", r.URL.Path)

		// if this is a banned name, just redirect to the local URI. That'll show em.
		if isBannedName(p) {
			http.Redirect(w, r, fmt.Sprintf("/%s", p), http.StatusTemporaryRedirect)
			return
		}

		serveAsset(w, r, "edit.html") //apparently this is like a redirect?
	})
	mux.HandleFunc("/links/", func(w http.ResponseWriter, r *http.Request) {
		getLinks(ctx, w, r)
	})
	mux.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request) {
		serveAsset(w, r, r.URL.Path[len("/s/"):])
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		serveAsset(w, r, "favicon.ico")
	})
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, version)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "👍")
	})

	// TODO(knorton): Remove the admin handler.
	if admin {
		mux.Handle("/admin/", &adminHandler{ctx})
	}

	return http.ListenAndServe(addr, mux)
}
