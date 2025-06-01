package internal

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Songmu/gitconfig"
	"github.com/TecharoHQ/yeet/internal/yeet"
)

var (
	GPGKeyFile      = flag.String("gpg-key-file", gpgKeyFileLocation(), "GPG key file to sign the package")
	GPGKeyID        = flag.String("gpg-key-id", "", "GPG key ID to sign the package")
	GPGKeyPassword  = flag.String("gpg-key-password", "", "GPG key password to sign the package")
	UserName        = flag.String("git-user-name", GitUserName(), "user name in Git")
	UserEmail       = flag.String("git-user-email", GitUserEmail(), "user email in Git")
	SourceDateEpoch = flag.Int64("source-date-epoch", GetSourceDateEpoch(), "Timestamp to use for all files in packages")
)

const (
	fallbackName  = "Mimi Yasomi"
	fallbackEmail = "mimi@xeserv.us"
)

func gpgKeyFileLocation() string {
	folder, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	return filepath.Join(folder, "techaro.lol", "yeet", "key.asc")
}

func GitUserName() string {
	name, err := gitconfig.User()
	if err != nil {
		return fallbackName
	}

	return name
}

func GitUserEmail() string {
	email, err := gitconfig.Email()
	if err != nil {
		return fallbackEmail
	}

	return email
}

func GitVersion() string {
	vers, err := yeet.GitTag(context.Background())
	if err != nil {
		panic(err)
	}

	return vers
}

func GetSourceDateEpoch() int64 {
	// fallback needs to be 1 because some software thinks unix time 0 means "no time"
	const fallback = 1

	gitPath, err := exec.LookPath("git")
	if err != nil {
		slog.Warn("git not found in $PATH", "err", err)
		return fallback
	}

	epochFromGitStr, err := yeet.Output(context.Background(), gitPath, "log", "-1", "--format=%ct")
	if err == nil {
		num, _ := strconv.ParseInt(strings.TrimSpace(epochFromGitStr), 10, 64)
		if num != 0 {
			return num
		}
	}

	return fallback
}

func SourceEpoch() time.Time {
	return time.Unix(*SourceDateEpoch, 0)
}
