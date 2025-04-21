package gitea

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

var (
	giteaHost     = flag.String("gitea-host", "", "URL of the gitea instance you are deploying to without a trailing slash, if not set then gitea integrations are no-ops")
	giteaToken    = flag.String("gitea-token", "", "Gitea token")
	giteaUsername = flag.String("gitea-username", "", "Gitea username")
)

func UploadPackage(ctx context.Context, c *http.Client, owner, distro, component, fname string) error {
	if *giteaHost == "" {
		slog.Debug("gitea config not set, bailing")
		return nil
	}

	kind := ""

	switch filepath.Ext(fname) {
	case ".deb":
		kind = "debian/pool"
	case ".rpm":
		kind = "rpm"
	default:
		slog.Debug("wrong package kind", "fname", fname)
		return nil
	}

	fin, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("can't open %s: %w", fname, err)
	}
	defer fin.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/api/packages/%s/%s/%s/%s/upload", *giteaHost, owner, kind, distro, component), fin)
	if err != nil {
		return fmt.Errorf("[unexpected] can't make request: %w", err)
	}

	req.SetBasicAuth(*giteaUsername, *giteaToken)

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("can't do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("got wrong status code from gitea: %s", resp.Status)
	}

	return nil
}
