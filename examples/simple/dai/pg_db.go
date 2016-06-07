package dai

import (
	"errors"

	"github.com/jimmy-go/pgwp"
	_ "github.com/lib/pq"
)

var (
	// Db pool is used for module users and friends.
	Db     *pgwp.Pool
	inited bool

	errInited     = errors.New("package already set")
	errOptionsNil = errors.New("invalid data for connection")
)

// Options struct
type Options struct {
	URL      string
	Workers  int
	Queue    int
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// Connect func
func Connect(opts *Options) error {
	if inited {
		return errInited
	}
	if opts == nil {
		return errOptionsNil
	}
	var err error
	Db, err = pgwp.Connect("postgres", opts.URL, opts.Workers, opts.Queue)
	if err != nil {
		return err
	}
	inited = true
	return err
}
