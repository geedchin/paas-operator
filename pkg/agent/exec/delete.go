package exec

import (
	"github.com/golang/glog"
)

func Delete(user, url string) error {
	fileName, err := wgetScript(url)
	if err != nil {
		glog.Fatalf("%s: No Such File Or Directory\n", url)
		return err
	}

	// args always is empty string
	return execScript(user, fileName, "")
}
