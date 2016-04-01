package main

import (
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/ipfs/go-ipfs-api"
	"github.com/meatballhat/negroni-logrus"
	"github.com/wmark/semver"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app         = kingpin.New("npm-ipfs", "Glue npm to a directory, providing semver resolution")
	ipfsAPI     = app.Flag("ipfs_api", "URL of IPFS API.").Envar("IPFS_API").Default("http://127.0.0.1:5001").URL()
	ipfsGateway = app.Flag("ipfs_gateway", "URL of IPFS API.").Envar("IPFS_GATEWAY").Default("http://127.0.0.1:8080").URL()
	addr        = app.Flag("addr", "Address to listen on.").Envar("ADDR").Default(":3001").String()
)

func getPackage(s *shell.Shell, r, n string) (pkg, error) {
	files, err := s.List("/ipns/" + r)
	if err != nil {
		return nil, err
	}

	var l pkg

	for _, f := range files {
		if !strings.HasPrefix(f.Name, n+"@") {
			continue
		}

		bits := strings.Split(f.Name, "@")

		vs := strings.TrimSuffix(bits[1], ".tgz")

		v, err := semver.NewVersion(vs)
		if err != nil {
			panic(err)
		}

		l = append(l, pkgVersion{
			hash:    f.Hash,
			name:    n,
			vstring: vs,
			version: v,
		})
	}

	sort.Reverse(l)

	return l, nil
}

type pkgVersion struct {
	hash    string
	name    string
	vstring string
	version *semver.Version
}

type pkg []pkgVersion

func (p pkg) Len() int           { return len(p) }
func (p pkg) Less(i, j int) bool { return p[i].version.Less(p[j].version) }
func (p pkg) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	sh := shell.NewShell((*ipfsAPI).String())

	m := mux.NewRouter()

	m.Methods("GET").Path("/{repo}/{name}@{spec}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		repo, name, spec := vars["repo"], vars["name"], vars["spec"]

		s, err := semver.NewRange(vars["spec"])
		if err != nil {
			panic(err)
		}

		logrus.WithFields(logrus.Fields{
			"repo": repo,
			"name": name,
			"spec": spec,
		}).Info("package request")

		l, err := getPackage(sh, repo, name)
		if err != nil {
			panic(err)
		}

		for _, p := range l {
			if s.IsSatisfiedBy(p.version) {
				logrus.WithFields(logrus.Fields{
					"name":    name,
					"spec":    spec,
					"version": p.vstring,
				}).Info("serving package")

				http.Redirect(w, r, (*ipfsGateway).ResolveReference(&url.URL{
					Path: "ipfs/" + p.hash,
				}).String(), http.StatusFound)

				return
			}
		}

		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	n := negroni.New()

	n.Use(negronilogrus.NewMiddleware())
	n.Use(negroni.NewRecovery())
	n.UseHandler(m)

	logrus.Info("listening")

	if err := http.ListenAndServe(*addr, n); err != nil {
		panic(err)
	}
}
