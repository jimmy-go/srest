package views

import "errors"

var (
	// Db pool is used for module users and friends.
	inited    bool
	errInited = errors.New("package already set")
)

// Configure func
func Configure(path string) error {
	if inited {
		return errInited
	}
	var err error
	if err != nil {
		return err
	}
	inited = true
	return err
}

// Render func
func Render(view string, data interface{}) error {
	// TODO; some templating
	return nil
}
