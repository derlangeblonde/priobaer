package cmdtest

import "fmt"

const port = 6666
const maxAgeDefault = 60 * 60 * 24

var localhost string = fmt.Sprintf("http://localhost:%d", port)
