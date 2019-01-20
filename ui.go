//+build ui

package groundcontrol

import "net/http"

var UI http.FileSystem = http.Dir("ui/build")
